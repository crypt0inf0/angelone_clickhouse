package utils

import (
    "time"
    "github.com/cenkalti/backoff/v4"
)

// NewExponentialBackoff creates a new exponential backoff configuration
func NewExponentialBackoff() *backoff.ExponentialBackOff {
    b := backoff.NewExponentialBackOff()
    b.InitialInterval = 1 * time.Second
    b.MaxInterval = 30 * time.Second
    b.MaxElapsedTime = 5 * time.Minute
    b.Multiplier = 2.0
    b.RandomizationFactor = 0.1
    return b
}
