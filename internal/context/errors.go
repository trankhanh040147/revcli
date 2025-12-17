package context

import "github.com/trankhanh040147/revcli/internal/filter"

// SecretsError represents an error when secrets are detected in code
type SecretsError struct {
	Matches []filter.SecretMatch
}

// Error implements the error interface
func (e SecretsError) Error() string {
	return "potential secrets detected in code. Use --force to proceed anyway"
}

// ErrSecretsDetected is a sentinel error for secrets detection
var ErrSecretsDetected = SecretsError{}

