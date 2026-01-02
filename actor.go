package axy

// Actor is the behavior contract for an axy actor.
//
// Typical usage is to implement an actor by embedding [Base] and optionally
// overriding lifecycle hooks (OnSpawn, OnMessage, etc.). The embedded [Base]
// provides default implementations for the message loop, cancellation, and
// helper methods like Send/Do/Go.
type Actor interface {
	base() *Base

	// Key returns the actor's stable identifier (if set).
	Key() string

	// Send enqueues a message to this actor.
	//
	// The sender is used for tracing/diagnostics and can be used by the receiver
	// to reply. Returns false if message is nil or the actor is already canceled.
	Send(message any, sender Reference) bool

	// Perception returns a wrapper around this actor that uses the given perceiver
	// as the "sender identity" for outgoing messages and exposes perceiver helpers
	// (Ctx/Do/Go).
	Perception(perceiver Actor) Perception

	// Cancel requests actor shutdown.
	Cancel()

	// OnSpawn is called at the beginning of the actor goroutine, before the loop starts.
	OnSpawn()

	// OnSpawned is called right after OnSpawn, still on the actor goroutine.
	OnSpawned()
	// Do(function func()) chan bool

	// OnMessage is called for every delivered message, on the actor goroutine.
	OnMessage(message any, sender Reference)

	// OnCancel is called once when cancellation is requested.
	OnCancel()

	// OnCanceled is called right after OnCancel.
	OnCanceled()

	// OnDestroy is called once when the actor goroutine is about to exit.
	OnDestroy()

	// OnDestroyed is called right after OnDestroy.
	OnDestroyed()
}
