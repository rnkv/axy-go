package axy

import "context"

type Perception struct {
	reference Reference
	perceiver Actor
}

func (p Perception) Key() string {
	return p.reference.Key()
}

func (p Perception) Send(message any) bool {
	return p.reference.Send(message, p.perceiver)
}

func (p Perception) Cancel() {
	p.reference.Cancel()
}

func (p Perception) Ctx() context.Context {
	return p.perceiver.base().Ctx()
}

func (p Perception) Do(callable func()) chan bool {
	return p.perceiver.base().Do(callable)
}

func (p Perception) Go(callable func()) {
	p.perceiver.base().Go(callable)
}
