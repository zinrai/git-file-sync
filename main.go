package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "Path to configuration file")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "git-file-sync - Sync files from Git repository\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Run with default config.yaml in current directory\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Specify custom config file\n")
		fmt.Fprintf(os.Stderr, "  %s -config=/path/to/config.yaml\n\n", os.Args[0])
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

	logger.Debug("loading configuration", "path", *configPath)

	config, err := LoadConfig(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := Sync(config, logger); err != nil {
		logger.Error("sync failed", "error", err)
		os.Exit(1)
	}
}
