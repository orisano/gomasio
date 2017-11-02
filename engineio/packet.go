package engineio

import "io"

type PacketType int

const (
	OPEN PacketType = iota
	CLOSE
	PING
	PONG
	MESSAGE
	UPGRADE
	NOOP

	INVALID PacketType = -1
)

type Packet struct {
	Type PacketType
	Body io.Reader
}
