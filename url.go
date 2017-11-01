package gomasio

import (
	"log"
	"net/url"
)

type URLOptions struct {
	Scheme string
	Path   string
	Query  url.Values
	Logger *log.Logger
}

type URLOption func(o *URLOptions)

func GetURL(host string, opts ...URLOption) (*url.URL, error) {
	query := make(url.Values)
	query.Set("EIO", "3")
	query.Set("transport", "websocket")

	options := &URLOptions{
		Scheme: "ws",
		Path:   "/socket.io/",
		Query:  query,
	}

	for _, opt := range opts {
		opt(options)
	}

	u := new(url.URL)
	u.Host = host
	u.Scheme = options.Scheme
	u.Path = options.Path
	u.RawQuery = options.Query.Encode()

	return u, nil
}

func WithSecure(o *URLOptions) {
	o.Scheme = "wss"
}

func SetQuery(key, value string) URLOption {
	return func(o *URLOptions) {
		o.Query.Set(key, value)
	}
}

func DelQuery(key string) URLOption {
	return func(o *URLOptions) {
		o.Query.Del(key)
	}
}

func WithPath(p string) URLOption {
	return func(o *URLOptions) {
		o.Path = p
	}
}
