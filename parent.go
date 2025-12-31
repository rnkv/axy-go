package axy

import "context"

type Parent struct {
	child Life
	ctx   context.Context
	queue chan<- any
}

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
	case p.queue <- Envelope{sender: p.child, message: message}:
		return true
	}
}
