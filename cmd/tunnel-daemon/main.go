package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"mmtunnel/internal/desktop/daemon"
	desktopruntime "mmtunnel/internal/desktop/runtime"
	"mmtunnel/internal/logging"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:19081", "loopback daemon listen address")
	profiles := flag.String("profiles", defaultProfilePath(), "desktop profile store path")
	logLevel := flag.String("log-level", "info", "daemon log level")
	flag.Parse()

	log, err := logging.New(*logLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize logging:", err)
		os.Exit(1)
	}
	api, err := daemon.NewPersistent(*profiles, func(provider desktopruntime.Provider) *desktopruntime.Manager {
		return desktopruntime.NewManager(provider, desktopruntime.NewClientRunner(), 500)
	})
	if err != nil {
		log.Error("load profiles", "error", err)
		os.Exit(1)
	}

	server := &http.Server{Addr: *listen, Handler: api.Handler()}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Info("desktop daemon starting", "listen", *listen, "profiles", *profiles)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("desktop daemon stopped", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func defaultProfilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "profiles.json"
	}
	return filepath.Join(dir, "mmtunnel", "profiles.json")
}

var _ = slog.LevelInfo
