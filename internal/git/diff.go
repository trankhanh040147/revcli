package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiffResult contains the extracted diff and file contents
type DiffResult struct {
	// RawDiff is the raw git diff output
	RawDiff string
	// ModifiedFiles maps file paths to their full content
	ModifiedFiles map[string]string
	// FilePaths is a list of all modified file paths
	FilePaths []string
}

// GetDiff extracts the git diff and reads modified file contents
func GetDiff(staged bool) (*DiffResult, error) {
	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return nil, err
	}

	// Get the raw diff
	rawDiff, err := getRawDiff(staged)
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}

	if rawDiff == "" {
		return nil, fmt.Errorf("no changes detected. Make sure you have uncommitted changes")
	}

	// Parse file paths from the diff
	filePaths := parseFilePaths(rawDiff)

	// Read file contents for modified files
	modifiedFiles := make(map[string]string)
	for _, path := range filePaths {
		content, err := readFileContent(path)
		if err != nil {
			// File might have been deleted, skip it
			continue
		}
		modifiedFiles[path] = content
	}

	return &DiffResult{
		RawDiff:       rawDiff,
		ModifiedFiles: modifiedFiles,
		FilePaths:     filePaths,
	}, nil
}

// checkGitRepo verifies we're inside a git repository
func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository (or any of the parent directories)")
	}
	return nil
}

// getRawDiff runs git diff and returns the output
func getRawDiff(staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	// Add unified diff format for better context
	args = append(args, "-U3")

	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git diff failed: %s", stderr.String())
	}

	return stdout.String(), nil
}

// parseFilePaths extracts file paths from the git diff output
func parseFilePaths(diff string) []string {
	var paths []string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()

		// Look for lines like "diff --git a/path/to/file b/path/to/file"
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				// Extract the "b/path" part and remove the "b/" prefix
				bPath := parts[3]
				if strings.HasPrefix(bPath, "b/") {
					path := strings.TrimPrefix(bPath, "b/")
					if !seen[path] {
						paths = append(paths, path)
						seen[path] = true
					}
				}
			}
		}
	}

	return paths
}

// readFileContent reads the full content of a file
func readFileContent(path string) (string, error) {
	// Get the git root directory
	rootDir, err := getGitRoot()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(rootDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(content), nil
}

// getGitRoot returns the root directory of the git repository
func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetGitRoot is exported for use by other packages
func GetGitRoot() (string, error) {
	return getGitRoot()
}

