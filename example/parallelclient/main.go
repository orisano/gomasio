package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"log"
	"time"

	"github.com/orisano/go-retry"
	"github.com/orisano/gomasio"
	"github.com/orisano/gomasio/engineio"
	"github.com/orisano/gomasio/socketio"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func run() error {
	var sec int
	flag.IntVar(&sec, "sec", 10, "time")

	var workers int
	flag.IntVar(&workers, "workers", 5, "workers")

	var duration int
	flag.IntVar(&duration, "duration", 500, "millisec")

	var host string
	flag.StringVar(&host, "host", "localhost:8080", "socket.io server addr")

	flag.Parse()

	u, err := gomasio.GetURL(host)
	if err != nil {
		return errors.Wrap(err, "failed to construct url")
	}
	ctx := context.Background()

	wait := make(chan struct{})

	ptm := socketio.NewPacketTypeMux()
	ptm.HandleFunc(socketio.CONNECT, func(sctx socketio.Context) {
		b := make([]byte, 6)
		rand.Read(b)
		id := base64.StdEncoding.EncodeToString(b)

		<-wait
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(duration) * time.Millisecond):
				sctx.Emit("/message", id)
			}
		}
	})
	h := socketio.OverEngineIO(ptm)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(sec)*time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < workers; i++ {
		var conn gomasio.Conn
		err := retry.Do(func() error {
			var err error
			conn, err = gomasio.NewConn(u.String(), gomasio.WithQueueSize(100))
			return err
		})
		if err != nil {
			return errors.Wrap(err, "failed to create connection")
		}
		eg.Go(func() error {
			return engineio.Connect(ctx, conn, h)
		})
	}

	close(wait)
	log.Print("start")
	<-ctx.Done()
	return eg.Wait()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
