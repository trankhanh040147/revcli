package filter

import (
	"regexp"
	"strings"
)

// IgnoredPatterns contains file patterns to ignore during review
var IgnoredPatterns = []string{
	"go.sum",
	"go.mod",
	"vendor/",
	"_generated.go",
	".pb.go",
	"_test.go",
	".mock.go",
	"mocks/",
	"testdata/",
	".git/",
	"node_modules/",
	"dist/",
	"build/",
}

// SecretPatterns contains regex patterns that might indicate secrets
var SecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["']?[a-zA-Z0-9_\-]{20,}["']?`),
	regexp.MustCompile(`(?i)(secret|password|passwd|pwd)\s*[:=]\s*["'][^"']{8,}["']`),
	regexp.MustCompile(`(?i)(token|bearer)\s*[:=]\s*["']?[a-zA-Z0-9_\-\.]{20,}["']?`),
	regexp.MustCompile(`(?i)private[_-]?key\s*[:=]`),
	regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
	regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key[_-]?id|aws[_-]?secret[_-]?access[_-]?key)\s*[:=]\s*["']?[A-Z0-9]{16,}["']?`),
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),                                     // GitHub personal access token
	regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),                                     // GitHub OAuth token
	regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),                                     // OpenAI API key
	regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),                                  // Google API key
	regexp.MustCompile(`(?i)database[_-]?url\s*[:=]\s*["']?[a-zA-Z]+://[^"'\s]+`), // Database URLs with credentials
}

// FilterResult contains the filtering results
type FilterResult struct {
	// FilteredFiles maps file paths to their content after filtering
	FilteredFiles map[string]string
	// IgnoredFiles lists files that were ignored
	IgnoredFiles []string
	// SecretsFound contains potential secrets that were detected
	SecretsFound []SecretMatch
}

// SecretMatch represents a potential secret found in the code
type SecretMatch struct {
	FilePath string
	Line     int
	Match    string
	Pattern  string
}

// Filter filters out ignored files and scans for secrets
func Filter(files map[string]string, rawDiff string) *FilterResult {
	result := &FilterResult{
		FilteredFiles: make(map[string]string),
		IgnoredFiles:  []string{},
		SecretsFound:  []SecretMatch{},
	}

	for path, content := range files {
		// Check if file should be ignored
		if shouldIgnore(path) {
			result.IgnoredFiles = append(result.IgnoredFiles, path)
			continue
		}

		// Scan for secrets
		secrets := scanForSecrets(path, content)
		result.SecretsFound = append(result.SecretsFound, secrets...)

		result.FilteredFiles[path] = content
	}

	// Also scan the raw diff for secrets
	diffSecrets := scanForSecrets("diff", rawDiff)
	result.SecretsFound = append(result.SecretsFound, diffSecrets...)

	return result
}

// shouldIgnore checks if a file path matches any ignored pattern
func shouldIgnore(path string) bool {
	for _, pattern := range IgnoredPatterns {
		// Check if pattern is a suffix match (for extensions)
		if strings.HasSuffix(pattern, ".go") || strings.HasSuffix(pattern, ".sum") || strings.HasSuffix(pattern, ".mod") {
			if strings.HasSuffix(path, pattern) || path == pattern {
				return true
			}
		}
		// Check if pattern is a directory prefix
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(path, pattern) || strings.Contains(path, "/"+pattern) {
				return true
			}
		}
		// Exact match
		if path == pattern {
			return true
		}
	}
	return false
}

// scanForSecrets scans content for potential secrets
func scanForSecrets(filePath, content string) []SecretMatch {
	var matches []SecretMatch

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		for _, pattern := range SecretPatterns {
			if match := pattern.FindString(line); match != "" {
				// Mask the actual secret value for safety
				maskedMatch := maskSecret(match)
				matches = append(matches, SecretMatch{
					FilePath: filePath,
					Line:     lineNum + 1,
					Match:    maskedMatch,
					Pattern:  pattern.String(),
				})
			}
		}
	}

	return matches
}

// maskSecret masks the sensitive part of a secret
func maskSecret(secret string) string {
	if len(secret) <= 10 {
		return "***REDACTED***"
	}

	// Show first 5 and last 3 characters
	prefix := secret[:5]
	suffix := secret[len(secret)-3:]
	return prefix + "***" + suffix
}

// HasSecrets returns true if any secrets were found
func (r *FilterResult) HasSecrets() bool {
	return len(r.SecretsFound) > 0
}

// FilterDiff filters the raw diff to remove ignored file changes
func FilterDiff(rawDiff string) string {
	var result strings.Builder
	var currentFile string
	var skipFile bool

	lines := strings.Split(rawDiff, "\n")
	for _, line := range lines {
		// Detect new file in diff
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				bPath := parts[3]
				if strings.HasPrefix(bPath, "b/") {
					currentFile = strings.TrimPrefix(bPath, "b/")
					skipFile = shouldIgnore(currentFile)
				}
			}
		}

		// Include line if we're not skipping this file
		if !skipFile {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

