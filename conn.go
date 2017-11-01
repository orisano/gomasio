package gomasio

import (
	"errors"
	"io"

	"github.com/gorilla/websocket"
)

// ref: https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
type Conn struct {
	conn *websocket.Conn
}

func NewConn(urlStr string) (*Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn: conn,
	}, nil
}

func (c *Conn) NextReader() (io.Reader, error) {
	mt, r, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	if mt != websocket.TextMessage {
		return nil, errors.New("support to text message only")
	}
	return r, nil
}

func (c *Conn) NextWriter() (io.WriteCloser, error) {
	return c.conn.NextWriter(websocket.TextMessage)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
