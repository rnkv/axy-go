package axy

// Spawn starts a new actor in the global system and returns its [Reference].
//
// For more control (e.g. scoping and waiting per system), create a [System] and
// use [System.Spawn] instead.
func Spawn(actor Actor) Reference {
	return globalSystem.spawn(actor, nil)
}
