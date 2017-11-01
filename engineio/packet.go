package engineio

import "io"

type PacketType int

const (
	Open PacketType = iota
	Close
	Ping
	Pong
	Message
	Upgrade
	Noop

	Invalid PacketType = -1
)

type Packet struct {
	Type PacketType
	Body io.Reader
}
