package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"OurBear/internal/service"
	"OurBear/internal/service/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	apiKey              = "7692071314:AAG4iOopqo6GAiZYiPbxIORKOCgBYzYYwUk" // да поебать мне на этот ключ ))))
	getUpdatedDelay     = 100 * time.Millisecond
	httpRequestsTimeout = 10 * time.Second
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.Stamp,
	}).Level(zerolog.DebugLevel)

	ctx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	service := service.New(config.Config{
		ApiKey:  apiKey,
		Delay:   getUpdatedDelay,
		Timeout: httpRequestsTimeout,
	})

	log.Info().Msg("starting service")
	if err := service.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("run error")
	}
}
