package socketio

import (
	"io"
)

type PacketType int

const (
	CONNECT PacketType = iota
	DISCONNECT
	EVENT
	ACK
	ERROR
	BINARY_EVENT
	BINARY_ACK
)

type Packet struct {
	Type        PacketType
	Attachments int
	Namespace   string
	ID          int
	Body        io.Reader
}
