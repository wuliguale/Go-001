package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	eg, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	eg.Go(func() error {
		var srv http.Server

		go func() {
			select {
			case <-ctx.Done():
				srv.Shutdown(context.Background())
			}
		}()

		return srv.ListenAndServe()
	})

	signalReceived := <-signalChan
	log.Printf("收到信号: %v，准备cancel", signalReceived)
	cancel()

	err := eg.Wait()
	if err != nil {
		log.Printf("errgroup err: %v", err)
	}

	log.Println("the end")
}
