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
	r, err := conn.NextReader()
	if err != nil {
		return errors.Wrap(err, "failed to get reader")
	}
	session, err := readHandshake(r)
	if err != nil {
		return errors.Wrap(err, "failed to read handshake data")
	}
	s := &socket{
		conn:         conn,
		pingInterval: time.Duration(session.PingInterval) * time.Millisecond,
		pingTimeout:  time.Duration(session.PingTimeout) * time.Millisecond,
		timeout:      make(chan struct{}),
	}
	defer s.Close()
	return listen(ctx, s, handler)
}

func readHandshake(r io.Reader) (*Session, error) {
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
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wf := NewWriterFactory(s.conn)
	s.PingAfter()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.timeout:
			return errors.New("timeout ping response")
		default:
			r, err := s.conn.NextReader()
			if err != nil {
				return errors.Wrap(err, "failed to get reader")
			}
			p, err := NewDecoder(r).Decode()
			if err != nil {
				return errors.Wrap(err, "failed to decode engine.io packet")
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
				wg.Add(1)
				go func() {
					defer wg.Done()
					handler.HandleMessage(wf, p.Body)
				}()
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

	timeout chan struct{}

	pingCancel context.CancelFunc

	timeoutLock   sync.Mutex
	timeoutCancel context.CancelFunc
}

func (s *socket) PingAfter() {
	if s.pingCancel != nil {
		s.pingCancel()
	}
	ctx := context.Background()
	ctx, s.pingCancel = context.WithCancel(ctx)

	t := time.NewTimer(s.pingInterval)
	go func() {
		defer t.Stop()
		select {
		case <-t.C:
			wf := s.conn.NewWriter()
			WritePing(wf)
			wf.Flush()
			s.setTimeout(s.pingTimeout)
		case <-ctx.Done():
		}
	}()
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
	ctx, s.timeoutCancel = context.WithCancel(ctx)

	t := time.NewTimer(d)
	go func() {
		defer t.Stop()
		select {
		case <-t.C:
			s.timeout <- struct{}{}
		case <-ctx.Done():
		}
	}()
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
