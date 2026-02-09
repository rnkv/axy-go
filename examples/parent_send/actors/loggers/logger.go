package loggers

import (
	"fmt"

	axy "github.com/rnkv/axy-go"
)

// Logger is a tiny actor that demonstrates "a view of another actor":
// other actors hold a Reference to it and send it events to persist/print/etc.
type Logger struct {
	axy.Base
}

func NewLogger() *Logger {
	l := &Logger{}
	return l
}

func (l *Logger) OnSpawn() {
	fmt.Println("Logger spawning...")
}

func (l *Logger) OnSpawned() {
	fmt.Println("Logger spawned.")
}

func (l *Logger) OnCanceled() {
	fmt.Println("Logger canceled.")
}

func (l *Logger) OnMessage(message any, sender axy.Reference) {
	switch message := message.(type) {
	case Text:
		fmt.Println("Logger got text:", message.Text, "sender:", sender.Key())
	default:
		// ignore
	}
}

func (l *Logger) OnDestroyed() {
	fmt.Println("Logger destroyed.")
}
