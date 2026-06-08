package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mmsocket/internal/config"
	"mmsocket/internal/logging"
	"mmsocket/internal/server"
)

func main() {
	configPath := flag.String("config", "server.yaml", "server YAML configuration path")
	flag.Parse()

	cfg, err := config.LoadServer(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load server config:", err)
		os.Exit(1)
	}
	log, err := logging.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize logging:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.New(cfg, log).Run(ctx); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
