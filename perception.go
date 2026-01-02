package axy

import "context"

// Perception is a convenience wrapper around a [Reference] scoped to a perceiver.
//
// It sends messages to the referenced actor while using the perceiver as the
// sender identity. It also exposes helpers (Ctx/Do/Go) from the perceiver,
// which makes it ergonomic to work with from within an actor.
type Perception struct {
	reference Reference
	perceiver Actor
}

// Key returns the referenced actor key.
func (p Perception) Key() string {
	return p.reference.Key()
}

// Send sends a message to the referenced actor using the perceiver as sender.
func (p Perception) Send(message any) bool {
	return p.reference.Send(message, p.perceiver)
}

// Cancel requests shutdown of the referenced actor.
func (p Perception) Cancel() {
	p.reference.Cancel()
}

// Ctx returns the perceiver's actor context (canceled during perceiver shutdown).
func (p Perception) Ctx() context.Context {
	return p.perceiver.base().Ctx()
}

// Do schedules callable to run on the perceiver's actor goroutine.
func (p Perception) Do(callable func()) chan bool {
	return p.perceiver.base().Do(callable)
}

// Go starts a perceiver-owned background goroutine tracked during shutdown.
func (p Perception) Go(callable func()) {
	p.perceiver.base().Go(callable)
}
