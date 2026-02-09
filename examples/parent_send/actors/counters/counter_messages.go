package counters

import "time"

// Protocol messages (what Counter accepts).
type Start struct {
	Every time.Duration
}

type Stop struct{}

// Outbound events (what Counter emits to its parent).
type Value struct {
	N int
}

// Destroyed is emitted when the counter has fully stopped and is destroyed.
type Destroyed struct{}
