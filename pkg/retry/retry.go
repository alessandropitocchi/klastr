package retry

import (
	"fmt"
	"strings"
	"time"
)

// transientPatterns are error substrings that indicate a transient/network error worth retrying.
var transientPatterns = []string{
	"timeout",
	"connection refused",
	"connection reset",
	"network",
	"i/o timeout",
	"no such host",
	"TLS handshake timeout",
	"server misbehaving",
	"dial tcp",
}

// IsTransient returns true if the error message matches known transient error patterns.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, pattern := range transientPatterns {
		if strings.Contains(msg, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// Run executes fn up to maxAttempts times, retrying only on transient errors.
// It uses exponential backoff starting from initialDelay.
// If logFn is non-nil, it is called with retry status messages.
func Run(maxAttempts int, initialDelay time.Duration, logFn func(string, ...any), fn func() error) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !IsTransient(lastErr) {
			return lastErr
		}

		if attempt < maxAttempts {
			if logFn != nil {
				logFn("Transient error (attempt %d/%d): %v. Retrying in %s...\n", attempt, maxAttempts, lastErr, delay)
			}
			time.Sleep(delay)
			delay *= 2 // exponential backoff
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
