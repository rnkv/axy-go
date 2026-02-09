package supervisors

import (
	"fmt"
	"time"

	axy "github.com/rnkv/axy-go"
	"github.com/rnkv/axy-go/examples/parent_send/actors/counters"
)

// Supervisor is the parent actor. It spawns the counter as a child and receives
// progress updates via Parent().Send(...) from the child.
type Supervisor struct {
	axy.Base
	counter *counter
	logger  *logger
}

func NewSupervisor() *Supervisor {
	return &Supervisor{}
}

func (s *Supervisor) OnSpawn() {
	fmt.Println("Supervisor spawning...")
	s.logger = newLogger(s)
	s.counter = newCounter(s)

	_ = s.counter.start(250 * time.Millisecond)

	// Stop after a bit.
	s.Go(func() {
		select {
		case <-s.Ctx().Done():
			return
		case <-time.After(1 * time.Second):
			<-s.Do(func() {
				_ = s.counter.stop()
			})
		}
	})
}

func (s *Supervisor) OnSpawned() {
	fmt.Println("Supervisor spawned.")
}

func (s *Supervisor) OnCanceled() {
	fmt.Println("Supervisor canceled.")
}

func (s *Supervisor) OnMessage(message any, sender axy.Reference) {
	switch message := message.(type) {
	case counters.Value:
		fmt.Println("Supervisor got counter value:", message.N)
		_ = s.logger.log(fmt.Sprintf("counter value: %d", message.N))
	case counters.Destroyed:
		// We cancel the supervisor when the counter is destroyed.
		s.Cancel()
	default:
		_ = sender
	}
}

func (s *Supervisor) OnDestroyed() {
	fmt.Println("Supervisor destroyed.")
}
