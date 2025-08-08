package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Sync fetches files from git repository and places them to destinations
func Sync(config *Config, logger *slog.Logger) error {
	logger.Info("starting sync",
		"repo", config.Git.RepoURL,
		"commit", config.Git.CommitHash,
		"files", len(config.Files))

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "git-sync-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Error("failed to cleanup temp directory", "path", tempDir, "error", err)
		}
	}()

	logger.Debug("created temp directory", "path", tempDir)

	// Clone repository with sparse-checkout
	repoDir := filepath.Join(tempDir, "repo")
	if err := cloneRepository(config, repoDir, logger); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Setup sparse-checkout
	if err := setupSparseCheckout(config, repoDir, logger); err != nil {
		return fmt.Errorf("failed to setup sparse-checkout: %w", err)
	}

	// Checkout specific commit
	if err := checkoutCommit(config, repoDir, logger); err != nil {
		return fmt.Errorf("failed to checkout commit: %w", err)
	}

	// Copy files to destinations
	if err := copyFilesToDestinations(config, repoDir, logger); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	logger.Info("sync completed successfully")
	return nil
}

func cloneRepository(config *Config, repoDir string, logger *slog.Logger) error {
	logger.Debug("cloning repository", "url", config.Git.RepoURL, "path", repoDir)

	// Build clone command with filter and no-checkout
	args := []string{
		"clone",
		"--filter=blob:none",
		"--no-checkout",
		config.Git.RepoURL,
		repoDir,
	}

	cmd := exec.Command("git", args...)

	// Set SSH command for authentication
	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null",
		config.SSHPrivateKeyPath)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	// Execute clone
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (output: %s)", err, string(output))
	}

	logger.Debug("repository cloned successfully")
	return nil
}

func setupSparseCheckout(config *Config, repoDir string, logger *slog.Logger) error {
	// Initialize sparse-checkout
	logger.Debug("initializing sparse-checkout")

	initCmd := exec.Command("git", "-C", repoDir, "sparse-checkout", "init")
	if output, err := initCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sparse-checkout init failed: %w (output: %s)", err, string(output))
	}

	// Collect unique paths (files and directories)
	pathSet := make(map[string]bool)
	for _, file := range config.Files {
		pathSet[file.Source] = true
	}

	// Convert to slice
	paths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		paths = append(paths, path)
	}

	logger.Debug("setting sparse-checkout paths", "paths", paths)

	// Set sparse-checkout paths
	args := []string{"-C", repoDir, "sparse-checkout", "set"}
	args = append(args, paths...)

	setCmd := exec.Command("git", args...)
	if output, err := setCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sparse-checkout set failed: %w (output: %s)", err, string(output))
	}

	return nil
}

func checkoutCommit(config *Config, repoDir string, logger *slog.Logger) error {
	logger.Debug("checking out commit", "commit", config.Git.CommitHash)

	cmd := exec.Command("git", "-C", repoDir, "checkout", config.Git.CommitHash)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w (output: %s)", err, string(output))
	}

	logger.Debug("checkout completed")
	return nil
}

func copyFilesToDestinations(config *Config, repoDir string, logger *slog.Logger) error {
	for _, file := range config.Files {
		src := filepath.Join(repoDir, file.Source)
		dst := file.Destination

		logger.Debug("copying file", "source", src, "destination", dst)

		// Check if source exists
		srcInfo, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("source not found: %s: %w", file.Source, err)
		}

		// Handle directory copy
		if srcInfo.IsDir() {
			if err := copyDirectory(src, dst, logger); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", file.Source, err)
			}
		} else {
			// Handle file copy
			// Create destination directory
			dstDir := filepath.Dir(dst)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", dstDir, err)
			}

			// Copy file
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", file.Source, err)
			}
		}

		logger.Info("copied", "source", file.Source, "destination", dst)
	}

	return nil
}

func copyDirectory(src, dst string, logger *slog.Logger) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Walk through source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip .git directory
		if strings.HasPrefix(relPath, ".git") {
			return filepath.SkipDir
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
