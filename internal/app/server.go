package app

import (
	"context"
	"faraway/config"
	"faraway/internal/server/tcp"
	"faraway/internal/usecases"
	"fmt"
	"log"
	"log/slog"
)

const (
	ErrPowInit   = "failed to initialize pow"
	ErrRunServer = "failed server run"
)

// RunServer started server application
func RunServer(ctx context.Context) error {
	cfg, err := config.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := slog.Default()
	logger = logger.With("Service", cfg.Name)

	powUsecase, err := usecases.NewPowUsecase(cfg.Pow.Difficulty)
	if err != nil {
		log.Fatal(ErrPowInit, err)
	}
	quoteUsecase := usecases.NewQuoteUsecase()

	server := tcp.NewServer(
		&tcp.Config{
			Address:    cfg.Server.Addr,
			KeepAlive:  cfg.Server.KeepAlive,
			Deadline:   cfg.Server.Deadline,
			BufferSize: 1024,
		},
		powUsecase,
		quoteUsecase,
		logger,
	)

	if err = server.Run(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
