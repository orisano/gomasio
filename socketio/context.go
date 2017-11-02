package socketio

import (
	"encoding/json"
	"io"

	"github.com/orisano/gomasio"
	"github.com/pkg/errors"
)

type Context interface {
	PacketType() PacketType
	Namespace() string
	Body() io.Reader

	Event() string
	Args(dst interface{}) error

	Emit(event string, args interface{}) error
	Disconnect() error
}

func NewContext(wf gomasio.WriterFactory, packet *Packet) (Context, error) {
	ctx := &context{
		wf:     wf,
		packet: packet,
	}
	if packet.Type == EVENT {
		var e Event
		if err := json.NewDecoder(packet.Body).Decode(&e); err != nil {
			return nil, errors.Wrap(err, "failed to decode event")
		}
		ctx.event = &e
	}
	return ctx, nil
}

type context struct {
	wf     gomasio.WriterFactory
	packet *Packet

	event *Event
}

func (c *context) PacketType() PacketType {
	return c.packet.Type
}

func (c *context) Namespace() string {
	return c.packet.Namespace
}

func (c *context) Body() io.Reader {
	return c.packet.Body
}

func (c *context) Event() string {
	return c.event.Name
}

func (c *context) Args(dst interface{}) error {
	return json.Unmarshal(c.event.Args, dst)
}

func (c *context) Emit(event string, args interface{}) error {
	e := &Event{
		Name: event,
	}
	if args != nil {
		b, err := json.Marshal(args)
		if err != nil {
			return errors.Wrap(err, "failed to marshal args")
		}
		e.Args = b
	}

	wf := c.wf.NewWriter()
	if _, err := wf.Write([]byte{byte(EVENT) + '0'}); err != nil {
		return errors.Wrap(err, "failed to write type")
	}
	if err := json.NewEncoder(wf).Encode(e); err != nil {
		return errors.Wrap(err, "failed to encode event")
	}
	return wf.Flush()
}

var disconnect = []byte{byte(DISCONNECT) + '0'}

func (c *context) Disconnect() error {
	wf := c.wf.NewWriter()
	if _, err := wf.Write(disconnect); err != nil {
		return err
	}
	return wf.Flush()
}
