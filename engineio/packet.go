package engineio

type PacketType int

const (
	Open PacketType = iota
	Close
	Ping
	Pong
	Message
	Upgrade
	Noop
)

type Packet struct {
	Type PacketType
	Data []byte
}
