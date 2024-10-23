package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"faraway/config"
	"faraway/internal/client/tcp"
	"faraway/internal/usecases"
)

// RunClient started client application
func RunClient(ctx context.Context) error {
	cfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger := slog.Default()
	logger = logger.With("Service", cfg.Name)

	solverUsecase, err := usecases.NewSolverUsecase(cfg.Difficulty)
	if err != nil {
		log.Fatal(ErrPowInit, err)
	}

	client := tcp.NewClient(
		&tcp.Config{
			ServerAddr:     cfg.ServerAddr,
			ConnectTimeout: 5 * time.Second,
			RequestTimeout: 5 * time.Second,
			RetryAttempts:  3,
			RetryDelay:     5 * time.Second,
			MaxMessageSize: 1024,
			BufferSize:     1024,
		},
		solverUsecase,
		logger,
	)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client: %w", err)
	}

	return nil
}
