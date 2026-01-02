package axy

import "context"

type Reference struct {
	key    string
	ctx    context.Context
	cancel context.CancelFunc
	queue  chan<- any
}

func (r Reference) Key() string {
	return r.key
}

func (r Reference) Send(message any, sender Reference) bool {
	if message == nil {
		return false
	}

	if r.ctx.Err() != nil {
		return false
	}

	select {
	case <-r.ctx.Done():
		return false
	case r.queue <- newEnvelope(sender, message):
		return true
	}
}

func (r Reference) Cancel() {
	r.cancel()
}
