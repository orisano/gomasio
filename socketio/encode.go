package socketio

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type Encoder struct {
	w *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: bufio.NewWriter(w),
	}
}

func (e *Encoder) Encode(packet *Packet) error {
	if packet == nil {
		return errors.New("missing packet")
	}
	e.w.WriteByte(byte(packet.Type) + '0')
	if len(packet.Namespace) > 0 && packet.Namespace != "/" {
		e.w.WriteString(packet.Namespace)
		e.w.WriteByte(',')
	}
	if packet.ID >= 0 {
		e.w.WriteString(strconv.Itoa(packet.ID))
	}
	if packet.Body != nil {
		io.Copy(e.w, packet.Body)
	}
	return e.w.Flush()
}
