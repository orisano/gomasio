package socketio

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	if err := e.w.WriteByte(byte(int(packet.Type) + '0')); err != nil {
		return err
	}
	if len(packet.Namespace) > 0 && packet.Namespace != "/" {
		if _, err := e.w.WriteString(packet.Namespace); err != nil {
			return err
		}
		if err := e.w.WriteByte(','); err != nil {
			return err
		}
	}
	if packet.ID >= 0 {
		if _, err := fmt.Fprint(e.w, packet.ID); err != nil {
			return err
		}
	}
	if packet.Data != nil {
		if _, err := e.w.Write(packet.Data); err != nil {
			return err
		}
	}
	if err := e.w.Flush(); err != nil {
		return err
	}
	return nil
}
