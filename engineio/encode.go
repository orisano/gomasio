package engineio

import (
	"bufio"
	"io"

	"golang.org/x/xerrors"
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
		return xerrors.New("missing packet")
	}
	e.w.WriteByte(byte(int(packet.Type) + '0'))
	if packet.Body != nil {
		io.Copy(e.w, packet.Body)
	}
	return e.w.Flush()
}

var ping = []byte{byte(PING) + '0'}

func WritePing(w io.Writer) error {
	_, err := w.Write(ping)
	return err
}
