package axy

// Cancel requests the global system to shutdown.
//
// It is safe to call multiple times.
func Cancel() {
	globalSystem.Cancel()
}
