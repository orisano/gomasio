package gomasio

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/orisano/gomasio/engineio"
)

type SocketIO struct {
	conn *websocket.Conn
}

func (s *SocketIO) Listen(ctx context.Context, ch chan<- *engineio.Packet) {
	for {
		select {
		case <-ctx.Done():
			return
		}
		mt, r, err := s.conn.NextReader()
		if err != nil {
			// logging err
			continue
		}
		if mt != websocket.TextMessage {
			// logging 'support to text message only'
			continue
		}
		ep, err := engineio.NewDecoder(r).Decode()
		if err != nil {
			// logging err
			continue
		}
		ch <- ep
	}
}
