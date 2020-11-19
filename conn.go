package gomasio

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/orisano/gomasio/internal"
)

type WriteFlusher interface {
	io.Writer
	Flush() error
}

type WriterFactory interface {
	NewWriter() WriteFlusher
}

type Conn interface {
	WriterFactory
	NextReader() (io.Reader, error)
	Close() error
}

// ref: https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
type conn struct {
	ws  *websocket.Conn
	wch chan io.Reader
}

type ConnOptions struct {
	QueueSize uint
	Header    http.Header
	Dialer    *websocket.Dialer
}

type ConnOption func(o *ConnOptions)

func WithQueueSize(qsize uint) ConnOption {
	return func(o *ConnOptions) {
		o.QueueSize = qsize
	}
}

func WithHeader(h http.Header) ConnOption {
	return func(o *ConnOptions) {
		o.Header = h
	}
}

func WithCookieJar(jar http.CookieJar) ConnOption {
	return func(o *ConnOptions) {
		o.Dialer.Jar = jar
	}
}

func NewConn(urlStr string, opts ...ConnOption) (Conn, error) {
	options := &ConnOptions{
		QueueSize: 100,
		Header:    nil,
		Dialer: &websocket.Dialer{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	for _, opt := range opts {
		opt(options)
	}

	ws, _, err := options.Dialer.Dial(urlStr, options.Header)
	if err != nil {
		return nil, err
	}

	wch := make(chan io.Reader, options.QueueSize)
	go func() {
		for r := range wch {
			wc, err := ws.NextWriter(websocket.TextMessage)
			if err != nil {
				internal.Log("[ERROR] failed to get writer: ", err)
				continue
			}
			if _, err := io.Copy(wc, r); err != nil {
				internal.Log("[ERROR] failed to write websocket: ", err)
				continue
			}
			wc.Close()
		}
	}()
	return &conn{
		ws:  ws,
		wch: wch,
	}, nil
}

func (c *conn) NextReader() (io.Reader, error) {
	mt, r, err := c.ws.NextReader()
	if err != nil {
		return nil, err
	}
	if mt != websocket.TextMessage {
		return nil, fmt.Errorf("currently supports only text message: %v", mt)
	}
	return r, nil
}

func (c *conn) NewWriter() WriteFlusher {
	return &asyncWriter{q: c.wch, buf: &bytes.Buffer{}}
}

func (c *conn) Close() error {
	close(c.wch)
	return c.ws.Close()
}

type asyncWriter struct {
	q   chan<- io.Reader
	buf *bytes.Buffer
}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func (w *asyncWriter) Flush() error {
	w.q <- w.buf
	return nil
}

type nopFlusher struct {
	w io.Writer
}

func (f *nopFlusher) Write(p []byte) (n int, err error) {
	return f.w.Write(p)
}

func (f *nopFlusher) Flush() error {
	return nil
}

func NopFlusher(w io.Writer) WriteFlusher {
	return &nopFlusher{w}
}
