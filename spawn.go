package axy

// Spawn starts a new actor in the global system and returns its [Reference].
//
// For more control (e.g. scoping and waiting per system), create a [System] and
// use [System.Spawn] instead.
func Spawn(actor Actor) Reference {
	reference := globalSystem.Spawn(actor)
	<-actor.base().onLive
	return reference
}

func spawn(actor Actor, parent *Base) Actor {
	base := actor.base()

	base.spawnOnce.Do(func() {
		base.actor = actor
		base.parent = parent
		base.onLive = make(chan struct{})

		if base.parent != nil {
			if len(base.parent.childActors) == 0 {
				base.parent.onChildActorsEmpty = make(chan struct{})
			}

			base.parent.childActors[actor] = struct{}{}
		}

		base.childActors = make(map[Actor]struct{})
		base.onChildActorsEmpty = make(chan struct{})
		close(base.onChildActorsEmpty)
		base.onDone = make(chan struct{})
		base.initializeQueue()
		base.initializeExternalCtx()
		base.initializeChildrenCtx()
		base.initializeInternalCtx()
		go base.live()
	})

	return actor
}
