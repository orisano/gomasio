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
	if p.Data != nil {
		if _, err := e.w.Write(p.Data); err != nil {
			return err
		}
	}
	return nil
}
