package axy

// Wait blocks until all actors spawned in the global system have exited.
func Wait() {
	globalSystem.Wait()
}
