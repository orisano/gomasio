package socketio

import (
	"bufio"
	"errors"
	"io"
	"strconv"

	"golang.org/x/xerrors"
)

var IllegalAttachmentsError = errors.New("illegal attachments")

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
		return nil, xerrors.Errorf("read first byte: %w", err)
	}
	x := b - '0'
	if x < 0 || 6 < x {
		return nil, xerrors.Errorf("invalid packet type(type=%v)", b)
	}

	p := &Packet{
		Type:        PacketType(x),
		Attachments: -1,
	}

	if p.Type == BINARY_EVENT || p.Type == BINARY_ACK {
		attachments, err := d.parseAttachments()
		if err != nil {
			return nil, xerrors.Errorf("parse attachments: %w", err)
		}
		p.Attachments = attachments
	}

	namespace, err := d.parseNamespace()
	if err != nil {
		return nil, xerrors.Errorf("parse namespace: %w", err)
	}
	p.Namespace = namespace

	id, err := d.parseID()
	if err != nil {
		return nil, xerrors.Errorf("parse ID: %w", err)
	}
	p.ID = id

	p.Body = d.r
	return p, nil
}

func (d *Decoder) parseAttachments() (int, error) {
	s, err := d.r.ReadString('-')
	if err == io.EOF {
		return -1, IllegalAttachmentsError
	}
	if err != nil {
		return -1, err
	}
	attachments, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return -1, IllegalAttachmentsError
	}
	return attachments, nil
}

func (d *Decoder) parseNamespace() (string, error) {
	b, err := d.r.ReadByte()
	if err == io.EOF {
		return "/", nil
	}
	if err != nil {
		return "", err
	}
	if b != '/' {
		d.r.UnreadByte()
		return "/", nil
	}
	s, err := d.r.ReadString(',')
	if err == io.EOF {
		return "/" + s, nil
	} else if err != nil {
		return "", err
	} else {
		return "/" + s[:len(s)-1], nil
	}
}

func (d *Decoder) parseID() (int, error) {
	id := -1
	for {
		b, err := d.r.ReadByte()
		if err == io.EOF {
			return id, nil
		}
		if err != nil {
			return -1, err
		}
		if b < '0' || '9' < b {
			d.r.UnreadByte()
			return id, nil
		}
		x := int(b - '0')
		if id == -1 {
			id = x
		} else {
			id = id*10 + x
		}
	}
}
