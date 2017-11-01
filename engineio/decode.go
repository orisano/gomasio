package engineio

import (
	"bufio"
	"fmt"
	"io"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (d *Decoder) Decode() (*Packet, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return nil, err
	}
	x := b - '0'
	if x < 0 || 6 < x {
		return nil, fmt.Errorf("invalid packet type. got: %v", b)
	}
	p := &Packet{
		Type: PacketType(x),
	}
	p.Body = d.r
	return p, nil
}

func ReadType(r io.Reader) (PacketType, error) {
	b := make([]byte, 1)
	if _, err := r.Read(b); err != nil {
		return Invalid, err
	}
	x := b[0] - '0'
	if x < 0 || 6 < x {
		return Invalid, fmt.Errorf("invalid packet type. got: %v", b[0])
	}
	return PacketType(x), nil
}
