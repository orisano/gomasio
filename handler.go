package gomasio

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/orisano/gomasio/engineio"
	"github.com/orisano/gomasio/socketio"
	"github.com/pkg/errors"
)

type Handler interface {
	HandlePacket(packet *socketio.Packet)
}

var TimeNow func() time.Time = time.Now

func Connect(ctx context.Context, conn *Conn, handler Handler) error {
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
	s.setPing()

	for {
		r, err := conn.NextReader()
		if err != nil {
			// logging err
			return err
		}
		ep, err := engineio.NewDecoder(r).Decode()
		if err != nil {
			// logging err
			continue
		}
		s.heartbeat()
		switch ep.Type {
		case engineio.Open:
			return errors.New("invalid communication flow")
		case engineio.Close:
			return nil
		case engineio.Ping:
			return errors.New("unexpected server ping")
		case engineio.Pong:
			s.setPing()
			break
		case engineio.Message:
			sp, err := socketio.NewDecoder(ep.Body).Decode()
			if err != nil {
				return errors.Wrap(err, "invalid socket.io packet")
			}

		case engineio.Upgrade:
			return errors.New("not support upgrade packet")
		case engineio.Noop:
			break
		}
	}
}

func handshake(conn *Conn) (*engineio.Session, error) {
	r, err := conn.NextReader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initial read")
	}
	ep, err := engineio.NewDecoder(r).Decode()
	if err != nil {
		return nil, errors.Wrap(err, "invalid initial engine.io packet")
	}
	if ep.Type != engineio.Open {
		return nil, errors.Errorf("unexpected engine.io packet type. expected: %v, but got: %v", engineio.Open, ept)
	}

	var session engineio.Session
	if err := json.NewDecoder(ep.Body).Decode(&session); err != nil {
		return nil, errors.Wrap(err, "invalid session json")
	}
	return &session, nil
}

type socket struct {
	conn         *Conn
	pingInterval time.Duration
	pingTimeout  time.Duration

	closed chan struct{}

	pingCtx    context.Context
	pingCancel context.CancelFunc

	timeoutLock   sync.Mutex
	timeoutCtx    context.Context
	timeoutCancel context.CancelFunc
}

func (s *socket) setPing() {
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
			wc, err := s.conn.NextWriter()
			if err != nil {
				// logging err
				return
			}
			defer wc.Close()
			if err := engineio.WritePing(wc); err != nil {
				// logging err
				return
			}
			s.setTimeout(s.pingTimeout)
		}
	}(s.pingCtx)
}

func (s *socket) heartbeat() {
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

func (s *socket) close() {
	if s.pingCancel != nil {
		s.pingCancel()
	}
	s.timeoutLock.Lock()
	defer s.timeoutLock.Unlock()
	if s.timeoutCancel != nil {
		s.timeoutCancel()
	}
}
