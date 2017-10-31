package gomasio

import (
	"io/ioutil"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/orisano/gomasio/socketio"
	"github.com/pkg/errors"
)

type Client struct {
	conn   *websocket.Conn
	logger *log.Logger
}

type ClientOptions struct {
	Scheme string
	Path   string
	Query  url.Values
	Logger *log.Logger
}

type ClientOption func(o *ClientOptions)

func NewClient(host string, opts ...ClientOption) (*Client, error) {
	query := make(url.Values)
	query.Set("EIO", "3")
	query.Set("transport", "websocket")

	options := &ClientOptions{
		Scheme: "ws",
		Path:   "/socket.io/",
		Query:  query,
		Logger: log.New(ioutil.Discard, "", log.Llongfile),
	}

	for _, opt := range opts {
		opt(options)
	}

	u := new(url.URL)
	u.Host = host
	u.Scheme = options.Scheme
	u.Path = options.Path
	u.RawQuery = options.Query.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect websocket")
	}

	return &Client{
		conn:   conn,
		logger: options.Logger,
	}, nil
}

func WithSecure(o *ClientOptions) {
	o.Scheme = "wss"
}

func SetQuery(key, value string) ClientOption {
	return func(o *ClientOptions) {
		o.Query.Set(key, value)
	}
}

func DelQuery(key string) ClientOption {
	return func(o *ClientOptions) {
		o.Query.Del(key)
	}
}

func WithPath(p string) ClientOption {
	return func(o *ClientOptions) {
		o.Path = p
	}
}

func WithLogger(logger *log.Logger) ClientOption {
	return func(o *ClientOptions) {
		o.Logger = logger
	}
}

func (c *Client) ReadSocketIO() (*socketio.Packet, error) {
	mt, r, err := c.conn.NextReader()
	if err != nil {
		return nil, err
	}
	if mt != websocket.TextMessage {
		return nil, errors.New("support text message only")
	}
	_ = r
	return nil, nil
}
