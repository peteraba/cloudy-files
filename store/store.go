package store

import "time"

// DefaultWaitTime is the default time to wait between retries.
var DefaultWaitTime = 100 * time.Millisecond

const (
	defaultMaxRetries  = 10
	defaultPermissions = 0o600
)
