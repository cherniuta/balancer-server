package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"test-assignment/balancer/adapters/rest"
	"test-assignment/balancer/adapters/rest/middleware"
	"test-assignment/balancer/config"
	"test-assignment/balancer/core"
	"test-assignment/balancer/core/limiter"

	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	serverPool := core.NewServerPool(log)

	addresses := []string{
		cfg.HelloConfig.FirstAddress,
		cfg.HelloConfig.SecondAddress,
		cfg.HelloConfig.ThirdAddress,
	}

	for _, addr := range addresses {
		serverPool.AddBackand(addr)
	}

	go healthCheck(serverPool)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	limiter := limiter.NewClientLimiter(ctx, cfg.RateLimit, cfg.RateTime)

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     middleware.Rate(log, limiter)(middleware.Concurrency(log, rest.HandleRequest(log, serverPool), int64(cfg.Concurrency))),
	}

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}

func healthCheck(serverPool *core.ServerPool) {
	t := time.NewTicker(time.Second * 20)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			serverPool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}
