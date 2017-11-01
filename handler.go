package gomasio

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/orisano/gomasio/engineio"
	"github.com/orisano/gomasio/socketio"
	"github.com/pkg/errors"
)

type Handler interface {
}

var TimeNow func() time.Time = time.Now

func Connect(ctx context.Context, conn *Conn, handler Handler) error {
	r, err := conn.NextReader()
	if err != nil {
		return errors.Wrap(err, "failed to initial read")
	}
	ep, err := engineio.NewDecoder(r).Decode()
	if err != nil {
		return errors.Wrap(err, "invalid initial engine.io packet")
	}
	if got := ep.Type; got != engineio.Open {
		return errors.Errorf("unexpected engine.io packet type. expected: %v, but got: %v", engineio.Open, got)
	}

	pingInterval := 30 * time.Second
	pingTimeout := 60 * time.Second
	if len(ep.Data) > 0 {
		var session engineio.Session
		if err := json.Unmarshal(ep.Data, &session); err != nil {
			return errors.Wrap(err, "invalid session json")
		}
		pingInterval = time.Duration(session.PingInterval) * time.Millisecond
		pingTimeout = time.Duration(session.PingTimeout) * time.Millisecond
	}

	quit := make(chan struct{})
	pong := make(chan struct{}, 8)
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w, err := conn.NextWriter()
				if err != nil {
					// logging err
					continue
				}
				if err := engineio.WritePing(w); err != nil {
					// logging err
					continue
				}
				w.Close()
				go func() {
					select {
					case <-pong:
						return
					case <-time.After(pingTimeout):
						quit <- struct{}{}
					}
				}()
			}
		}
	}()

	for {
		r, err := conn.NextReader()
		if err != nil {
			return err
		}
		ep, err := engineio.NewDecoder(r).Decode()
		if err != nil {
			// logging err
			continue
		}
		switch ep.Type {
		case engineio.Open:
			return errors.New("invalid communication flow")
		case engineio.Close:
			return nil
		case engineio.Ping:
			return errors.New("unexpected server ping")
		case engineio.Pong:
			pong <- struct{}{}
			break
		case engineio.Message:
			sp, err := socketio.NewDecoder(bytes.NewReader(ep.Data)).Decode()
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
