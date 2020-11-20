package socketio

import (
	stdctx "context"
	"io"

	"github.com/orisano/gomasio"
	"github.com/orisano/gomasio/engineio"
)

type Handler interface {
	HandleSocketIO(ctx Context)
}

type HandleFunc func(ctx Context)

func (f HandleFunc) HandleSocketIO(ctx Context) {
	f(ctx)
}

type engineioHandler struct {
	handler Handler
}

func (h *engineioHandler) HandleMessage(wf gomasio.WriterFactory, body io.Reader) {
	p, err := NewDecoder(body).Decode()
	if err != nil && err != io.EOF {
		return
	}
	ctx, err := NewContext(wf, p)
	if err != nil {
		return
	}
	h.handler.HandleSocketIO(ctx)
}

func Connect(ctx stdctx.Context, conn gomasio.Conn, handler Handler) error {
	return engineio.Connect(ctx, conn, OverEngineIO(handler))
}

func OverEngineIO(handler Handler) engineio.Handler {
	return &engineioHandler{handler}
}

type EventMux struct {
	handlers map[string]Handler
}

func (m *EventMux) HandleFunc(event string, handleFunc func(ctx Context)) {
	m.Handle(event, HandleFunc(handleFunc))
}

func (m *EventMux) Handle(event string, handler Handler) {
	m.handlers[event] = handler
}

func (m *EventMux) HandleSocketIO(ctx Context) {
	ev := ctx.Event()
	handler, ok := m.handlers[ev]
	if ok {
		handler.HandleSocketIO(ctx)
	}
}

func NewEventMux() *EventMux {
	return &EventMux{
		handlers: make(map[string]Handler),
	}
}

type NamespaceMux struct {
	handlers map[string]Handler
}

func (m *NamespaceMux) Handle(namespace string, handler Handler) {
	m.handlers[namespace] = handler
}

func (m *NamespaceMux) HandleFunc(namespace string, handler func(ctx Context)) {
	m.Handle(namespace, HandleFunc(handler))
}

func (m *NamespaceMux) HandleSocketIO(ctx Context) {
	ns := ctx.Namespace()
	handler, ok := m.handlers[ns]
	if ok {
		handler.HandleSocketIO(ctx)
	}
}

func NewNamespaceMux() *NamespaceMux {
	return &NamespaceMux{
		handlers: make(map[string]Handler),
	}
}

type PacketTypeMux struct {
	handlers map[PacketType]Handler
}

func (m *PacketTypeMux) Handle(packetType PacketType, handler Handler) {
	m.handlers[packetType] = handler
}

func (m *PacketTypeMux) HandleFunc(packetType PacketType, handler func(ctx Context)) {
	m.Handle(packetType, HandleFunc(handler))
}

func (m *PacketTypeMux) HandleSocketIO(ctx Context) {
	pt := ctx.PacketType()
	handler, ok := m.handlers[pt]
	if ok {
		handler.HandleSocketIO(ctx)
	}
}

func NewPacketTypeMux() *PacketTypeMux {
	return &PacketTypeMux{
		handlers: make(map[PacketType]Handler),
	}
}
