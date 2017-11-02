package engineio

import (
	"github.com/orisano/gomasio"
)

func NewWriter(wf gomasio.WriteFlusher, packetType PacketType) gomasio.WriteFlusher {
	return gomasio.NewPrefixWriter(wf, []byte{byte(packetType) + '0'})
}

type writerFactory struct {
	wf gomasio.WriterFactory
}

func (w *writerFactory) NewWriter() gomasio.WriteFlusher {
	return NewWriter(w.wf.NewWriter(), MESSAGE)
}

func NewWriterFactory(wf gomasio.WriterFactory) gomasio.WriterFactory {
	return &writerFactory{
		wf: wf,
	}
}
