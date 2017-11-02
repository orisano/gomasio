package main

import (
	"context"
	"log"
	"time"

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

	ptm := socketio.NewPacketTypeMux()
	ptm.HandleFunc(socketio.CONNECT, func(ctx socketio.Context) {
		go func() {
			for i := 0; i < 30; i++ {
				ctx.Emit("count", i)
				time.Sleep(1 * time.Second)
			}
			ctx.Disconnect()
		}()
	})

	em := socketio.NewEventMux()
	em.HandleFunc("news", func(ctx socketio.Context) {
		var msg map[string]string
		ctx.Args(&msg)
		log.Print(msg)
	})
	ptm.Handle(socketio.EVENT, em)

	ctx := context.Background()
	return engineio.Connect(ctx, conn, socketio.OverEngineIO(ptm))
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
