package engineio

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/orisano/gomasio"
	"github.com/pkg/errors"
)

type Handler interface {
	HandleMessage(wf gomasio.WriterFactory, body io.Reader)
}

type HandleFunc func(wf gomasio.WriterFactory, body io.Reader)

func (f HandleFunc) HandleMessage(wf gomasio.WriterFactory, body io.Reader) {
	f(wf, body)
}

func Connect(ctx context.Context, conn gomasio.Conn, handler Handler) error {
	session, err := handshake(conn)
	if err != nil {
		return errors.Wrap(err, "failed to handshake")
	}
	s := &socket{
		conn:         conn,
		pingInterval: time.Duration(session.PingInterval) * time.Millisecond,
		pingTimeout:  time.Duration(session.PingTimeout) * time.Millisecond,
		closed:       make(chan struct{}),
	}
	defer s.Close()
	return listen(ctx, s, handler)
}

func handshake(conn gomasio.Conn) (*Session, error) {
	r, err := conn.NextReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initial read")
	}
	p, err := NewDecoder(r).Decode()
	if err != nil {
		return nil, errors.Wrap(err, "invalid initial engine.io packet")
	}
	if p.Type != OPEN {
		return nil, errors.Errorf("unexpected engine.io packet type. expected: %v, but got: %v", OPEN, p.Type)
	}

	var session Session
	if err := json.NewDecoder(p.Body).Decode(&session); err != nil {
		return nil, errors.Wrap(err, "invalid session json")
	}
	return &session, nil
}

func listen(ctx context.Context, s *socket, handler Handler) error {
	wf := NewWriterFactory(s.conn)
	for {
		select {
		case <-ctx.Done():
			// TODO: stop all spawned handlers
			return nil
		case <-s.closed:
			return errors.New("timeout ping")
		default:
			r, err := s.conn.NextReader()
			if err != nil {
				return err
			}
			p, err := NewDecoder(r).Decode()
			if err != nil {
				return err
			}
			s.Heartbeat()
			switch p.Type {
			case OPEN:
				return errors.New("invalid communication flow")
			case CLOSE:
				return nil
			case PING:
				return errors.New("unexpected server ping")
			case PONG:
				s.PingAfter()
				break
			case MESSAGE:
				go handler.HandleMessage(wf, p.Body)
			case UPGRADE:
				return errors.New("not support upgrade packet")
			case NOOP:
				break
			}
		}
	}
}

type socket struct {
	conn         gomasio.Conn
	pingInterval time.Duration
	pingTimeout  time.Duration

	closed chan struct{}

	pingCtx    context.Context
	pingCancel context.CancelFunc

	timeoutLock   sync.Mutex
	timeoutCtx    context.Context
	timeoutCancel context.CancelFunc
}

func (s *socket) PingAfter() {
	if s.pingCancel != nil {
		s.pingCancel()
	}
	ctx := context.Background()
	s.pingCtx, s.pingCancel = context.WithCancel(ctx)
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.pingInterval):
			wf := s.conn.NewWriter()
			defer wf.Flush()
			WritePing(wf)
			s.setTimeout(s.pingTimeout)
		}
	}(s.pingCtx)
}

func (s *socket) Heartbeat() {
	s.setTimeout(s.pingInterval + s.pingTimeout)
}

func (s *socket) setTimeout(d time.Duration) {
	s.timeoutLock.Lock()
	defer s.timeoutLock.Unlock()
	if s.timeoutCancel != nil {
		s.timeoutCancel()
	}
	ctx := context.Background()
	s.timeoutCtx, s.timeoutCancel = context.WithCancel(ctx)
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
			s.closed <- struct{}{}
		}
	}(s.timeoutCtx)
}

func (s *socket) Close() {
	if s.pingCancel != nil {
		s.pingCancel()
	}
	s.timeoutLock.Lock()
	defer s.timeoutLock.Unlock()
	if s.timeoutCancel != nil {
		s.timeoutCancel()
	}
}
