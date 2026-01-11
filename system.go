package axy

// System owns a set of actors and can be used to wait until they all exit.
//
// If you don't need explicit scoping, you can use the package-level [Spawn] and
// [Wait] which use a global system instance.
type system struct {
	Base
}

func newSystem() *system {
	return &system{}
}

type System interface {
	base() *Base
	Spawn(actor Actor) Reference
	Wait()
	Done() <-chan struct{}
}

// NewSystem creates an isolated actor system.
func NewSystem() System {
	return newSystem()
}

var globalSystem = newSystem()

func (s *system) OnSpawn()                                {}
func (s *system) OnSpawned()                              {}
func (s *system) OnMessage(message any, sender Reference) {}
func (s *system) OnCancel()                               {}
func (s *system) OnCanceled()                             {}
func (s *system) OnDestroy()                              {}
func (s *system) OnDestroyed()                            {}

// Spawn spawns a new actor in the system and returns a reference to it.
func (s *system) Spawn(actor Actor) Reference {
	spawn(s, nil)
	<-s.onLive

	onReference := make(chan Reference, 1)

	<-s.Do(func() {
		onReference <- s.base().Spawn(actor)
	})

	reference := <-onReference
	<-actor.base().onLive
	return reference
}

// Wait blocks until all actors spawned in this system have destroyed.
func (s *system) Wait() {
	spawn(s, nil)
	<-s.onDone
}

// Done returns a channel that is closed when all actors spawned in this system have destroyed.
func (s *system) Done() <-chan struct{} {
	spawn(s, nil)
	return s.onDone
}
