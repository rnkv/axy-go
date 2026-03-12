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

	// Spawn spawns a new actor in the system and returns a reference to it.
	Spawn(actor Actor) Reference

	// Wait blocks until all actors spawned in the system have destroyed.
	Wait()

	// Done returns a channel that is closed when the system has shutdown.
	Done() <-chan struct{}

	// Cancel requests the system to shutdown.
	//
	// It is safe to call multiple times.
	Cancel()
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

func (s *system) Wait() {
	spawn(s, nil)
	<-s.onDone
}

func (s *system) Done() <-chan struct{} {
	spawn(s, nil)
	return s.onDone
}
