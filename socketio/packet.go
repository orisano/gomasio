package socketio

type PacketType int

const (
	Connect PacketType = iota
	Disconnect
	Event
	Ack
	Error
	BinaryEvent
	BinaryAck
)

type Packet struct {
	Type        PacketType
	Attachments int
	Namespace   string
	ID          int
	Data        []byte
}
