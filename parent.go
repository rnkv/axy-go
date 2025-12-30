package axy

import "context"

type Parent struct {
	child Life
	ctx   context.Context
	queue chan<- any
}

func (p Parent) Send(message any) {
	if message == nil {
		return
	}

	if p.ctx.Err() != nil {
		return
	}

	select {
	case <-p.ctx.Done():
		return
	case p.queue <- Envelope{sender: p.child, message: message}:
		return
	}
}
