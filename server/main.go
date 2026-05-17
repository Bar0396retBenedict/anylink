package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
	// BuildDate is set at build time via ldflags
	BuildDate = "unknown"
)

func main() {
	var (
		confFile string
		showVersion bool
	)

	flag.StringVar(&confFile, "c", "conf/server.toml", "config file path")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("AnyLink Server\nVersion: %s\nBuildDate: %s\n", Version, BuildDate)
		os.Exit(0)
	}

	// Initialize logger
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetOutput(os.Stdout)

	logrus.Infof("Starting AnyLink Server v%s (built %s)", Version, BuildDate)

	// Load configuration
	cfg, err := LoadConfig(confFile)
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	if cfg.LogLevel != "" {
		level, err := logrus.ParseLevel(cfg.LogLevel)
		if err != nil {
			logrus.Warnf("Invalid log level %q, defaulting to info", cfg.LogLevel)
		} else {
			logrus.SetLevel(level)
		}
	}

	// Initialize and start the server
	srv, err := NewServer(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize server: %v", err)
	}

	if err := srv.Start(); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}

	logrus.Infof("AnyLink Server started, listening on %s", cfg.ServerAddr)

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")
	if err := srv.Stop(); err != nil {
		logrus.Errorf("Error during shutdown: %v", err)
	}
	logrus.Info("Server stopped")
}
