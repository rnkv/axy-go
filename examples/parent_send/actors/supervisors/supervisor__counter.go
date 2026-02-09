package supervisors

import (
	"time"

	axy "github.com/rnkv/axy-go"
	"github.com/rnkv/axy-go/examples/parent_send/actors/counters"
)

// counter is a first-class domain entity that represents the supervisor's
// relationship to a counter actor.
//
// It is intentionally private to the `supervisors` package: it encodes local policy
// and hides raw message sending behind intention-revealing helpers.
type counter struct {
	axy.Reference
	supervisor *Supervisor
}

func newCounter(supervisor *Supervisor) *counter {
	return &counter{
		Reference:  supervisor.Spawn(counters.NewCounter()),
		supervisor: supervisor,
	}
}

func (c *counter) start(every time.Duration) bool {
	if c == nil || c.Reference == nil {
		return false
	}

	// We as supervisor send the start command to the counter actor.
	return c.Send(counters.Start{
		Every: every,
	}, c.supervisor)
}

func (c *counter) stop() bool {
	if c == nil || c.Reference == nil {
		return false
	}

	// We as supervisor send the stop command to the counter actor.
	return c.Send(counters.Stop{}, c.supervisor)
}
