package engineio

import (
	"bufio"
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

func (e *Encoder) Encode(p *Packet) error {
	if err := e.w.WriteByte(byte(int(p.Type) + '0')); err != nil {
		return err
	}
	if len(p.Data) > 0 {
		if _, err := e.w.Write(p.Data); err != nil {
			return err
		}
	}
	return nil
}

var ping = []byte{byte('0' + int(Ping))}

func WritePing(w io.Writer) error {
	_, err := w.Write(ping)
	return err
}
