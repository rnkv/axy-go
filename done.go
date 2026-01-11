package axy

// Done returns a channel that is closed when all actors spawned in the global system have destroyed.
func Done() <-chan struct{} {
	return globalSystem.Done()
}
