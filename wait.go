package axy

// Wait blocks until all actors spawned in the global system have destroyed.
func Wait() {
	globalSystem.Wait()
}
