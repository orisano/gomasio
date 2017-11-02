package main

import (
	"context"
	"log"

	"github.com/orisano/gomasio"
	"github.com/orisano/gomasio/engineio"
	"github.com/orisano/gomasio/socketio"
	"github.com/pkg/errors"
)

func run() error {
	u, err := gomasio.GetURL("localhost:8080")
	if err != nil {
		return errors.Wrap(err, "failed to construct url")
	}
	conn, err := gomasio.NewConn(u.String(), 1000)
	if err != nil {
		return errors.Wrap(err, "failed to construct connection")
	}

	em := socketio.NewEventMux()
	em.HandleFunc("say", func(ctx socketio.Context) {
		ctx.Emit("reply", "client")
	})
	em.HandleFunc("reply", func(ctx socketio.Context) {
		ctx.Emit("say", "client")
	})
	ptm := socketio.NewPacketTypeMux()
	ptm.Handle(socketio.EVENT, em)

	ctx := context.Background()
	return engineio.Connect(ctx, conn, socketio.OverEngineIO(ptm))
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
