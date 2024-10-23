package main

import (
	"context"
	"faraway/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	go func() {
		<-ctx.Done()
	}()

	if err := app.RunServer(ctx); err != nil {
		log.Fatalf("failed to run client: %v", err)
	}
}
