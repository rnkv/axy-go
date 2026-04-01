package axy

// Reference is a handle to an actor.
//
// References are safe to share between goroutines and are the primary public way
// to interact with actors after spawning them.
type Reference interface {
	// Key returns the actor key (if any).
	Key() string

	// Send enqueues a message to the referenced actor.
	//
	// The sender is used for tracing/diagnostics and can be used by the receiver
	// to reply.
	// If the sender is not provided, the parent actor of the referenced actor
	// is used as the sender.
	// Returns false if message is nil or the actor is already canceled.
	Send(message any, sender ...Reference) bool

	// // Perception returns a wrapper that sends messages "as" the given perceiver.
	// Perception(perceiver Actor) Perception

	// Cancel requests shutdown of the referenced actor.
	Cancel()
}
