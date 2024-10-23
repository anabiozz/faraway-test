package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"faraway/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	go func() {
		<-ctx.Done()
	}()

	if err := app.RunClient(ctx); err != nil {
		log.Fatalf("failed to run client: %v", err)
	}
}
