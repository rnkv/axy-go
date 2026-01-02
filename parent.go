package axy

import "context"

// Parent is a handle for a child actor to send messages to its parent actor.
//
// It is produced by [Base.Parent]. Messages are sent with the child actor as
// the sender.
type Parent struct {
	child Actor
	ctx   context.Context
	queue chan<- any
}

// Send enqueues a message to the parent actor.
//
// Returns false if message is nil or the parent actor is already shutting down.
func (p Parent) Send(message any) bool {
	if message == nil {
		return false
	}

	if p.ctx.Err() != nil {
		return false
	}

	select {
	case <-p.ctx.Done():
		return false
	case p.queue <- envelope{sender: p.child, message: message}:
		return true
	}
}
