package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"mmtunnel/internal/client"
	"mmtunnel/internal/config"
	desktopruntime "mmtunnel/internal/desktop/runtime"
	"mmtunnel/internal/logging"
)

func main() {
	configPath := flag.String("config", "client.yaml", "client YAML configuration path")
	logLevel := flag.String("log-level", "info", "client log level")
	flag.Parse()

	cfg, err := config.LoadClient(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load client config:", err)
		os.Exit(1)
	}
	log, err := logging.New(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize logging:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runner := func(ctx context.Context, cfg config.ClientConfig, emit func(desktopruntime.Event)) error {
		return client.New(cfg, log).Run(ctx)
	}
	manager := desktopruntime.NewManager(desktopruntime.StaticProvider{ProfileID: "cli", Config: cfg}, runner, 200)
	if err := manager.Start(ctx); err != nil {
		log.Error("client failed to start", "error", err)
		os.Exit(1)
	}
	<-ctx.Done()
	_ = manager.Stop()
	for _, event := range manager.Events() {
		if event.Type == desktopruntime.EventError {
			log.Error("client runtime error", "error", event.Error)
		}
	}
}
