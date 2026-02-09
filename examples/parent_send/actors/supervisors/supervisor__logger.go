package supervisors

import (
	axy "github.com/rnkv/axy-go"
	"github.com/rnkv/axy-go/examples/parent_send/actors/loggers"
)

// logger is a first-class domain entity that represents the supervisor's
// relationship to a logger actor.
//
// It is intentionally private to the `supervisors` package.
type logger struct {
	axy.Reference
	supervisor *Supervisor
}

func newLogger(supervisor *Supervisor) *logger {
	return &logger{
		Reference:  supervisor.Spawn(loggers.NewLogger()),
		supervisor: supervisor,
	}
}

func (l *logger) log(text string) bool {
	if l == nil || l.Reference == nil {
		return false
	}

	// We as supervisor send the text to the logger actor.
	return l.Send(loggers.Text{
		Text: text,
	}, l.supervisor)
}
