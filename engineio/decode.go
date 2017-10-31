package engineio

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
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
	data, err := ioutil.ReadAll(d.r)
	if err != nil && err != io.EOF {
		return nil, err
	}
	p.Data = data
	return p, nil
}
