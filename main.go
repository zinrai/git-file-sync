package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "Path to configuration file")
		daemon     = flag.Bool("daemon", false, "Run in daemon mode")
		interval   = flag.Duration("interval", 60*time.Second, "Sync interval in daemon mode")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "git-file-sync - Sync files from Git repository\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Run once with default config\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Run with custom config file\n")
		fmt.Fprintf(os.Stderr, "  %s -config=/path/to/config.yaml\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Run in daemon mode, sync every 5 minutes\n")
		fmt.Fprintf(os.Stderr, "  %s -daemon -interval=5m\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Enable verbose logging\n")
		fmt.Fprintf(os.Stderr, "  %s -verbose\n", os.Args[0])
	}

	flag.Parse()

	// Setup logger
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Main sync function
	run := func() error {
		logger.Debug("loading configuration", "path", *configPath)

		config, err := LoadConfig(*configPath)
		if err != nil {
			return err
		}

		return Sync(config, logger)
	}

	if *daemon {
		logger.Info("starting in daemon mode", "interval", *interval)

		// Run once immediately
		if err := run(); err != nil {
			logger.Error("sync failed", "error", err)
		}

		// Then run periodically
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := run(); err != nil {
				logger.Error("sync failed", "error", err)
			}
		}
	} else {
		// One-shot mode
		if err := run(); err != nil {
			logger.Error("sync failed", "error", err)
			os.Exit(1)
		}
	}
}
