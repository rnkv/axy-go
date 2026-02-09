package counters

import (
	"fmt"
	"time"

	axy "github.com/rnkv/axy-go"
)

type Counter struct {
	axy.Base
	n int
}

func NewCounter() *Counter {
	c := &Counter{}
	return c
}

func (c *Counter) OnSpawn() {
	fmt.Println("Counter spawning...")
}

func (c *Counter) OnSpawned() {
	fmt.Println("Counter spawned.")
}

func (c *Counter) OnCanceled() {
	fmt.Println("Counter canceled.")
}

func (c *Counter) OnMessage(message any, sender axy.Reference) {
	switch message := message.(type) {
	case Start:
		// Background work is allowed, but it must not "act like an actor".
		// Hop back into the actor loop via Do(), then send events from there.
		c.Go(func() {
			ticker := time.NewTicker(message.Every)
			defer ticker.Stop()

			for {
				select {
				case <-c.Ctx().Done():
					return
				case <-ticker.C:
					<-c.Do(func() {
						c.n++

						_ = c.Parent().Send(Value{
							N: c.n,
						})
					})
				}
			}
		})
	case Stop:
		c.Cancel()
	default:
		_ = sender
	}
}

func (c *Counter) OnDestroyed() {
	fmt.Println("Counter destroyed.")

	// We send a destroyed event to the parent when the counter is destroyed.
	_ = c.Parent().Send(Destroyed{})
}
