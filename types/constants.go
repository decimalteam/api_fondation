package types

import "time"

const (
	RequestTimeout    = 16 * time.Second
	RequestRetryDelay = 32 * time.Millisecond
)
