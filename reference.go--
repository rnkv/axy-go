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

func (r Reference) Send(message any) {
	if message == nil {
		return
	}

	if r.ctx.Err() != nil {
		return
	}

	select {
	case <-r.ctx.Done():
		return
	case r.queue <- message:
		return
	}
}

func (r Reference) Cancel() {
	r.cancel()
}
