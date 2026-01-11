//go:build debug

package axy

import (
	"fmt"
	"runtime"
)

func goid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)

	// "goroutine 123 [running]:\n"
	var id uint64
	_, _ = fmt.Sscanf(string(buf[:n]), "goroutine %d ", &id)
	return id
}

// AssertInside panics if called outside the actor goroutine (debug only).
func (b *Base) AssertInside() {
	if b.goid != goid() {
		panic(
			fmt.Sprintf(
				"axy: contract violation: method must be called from the actor %s (%s) goroutine; "+
					"use Do() to hop to the actor loop",
				b.kind(),
				b.key,
			),
		)
	}
}

// AssertOutside panics if called inside the actor goroutine (debug only).
func (b *Base) AssertOutside() {
	if b.goid == goid() {
		panic(
			fmt.Sprintf(
				"axy: contract violation: method must be called from the outside of the actor %s (%s) goroutine; "+
					"use Go() to run a new child goroutine",
				b.kind(),
				b.key,
			),
		)
	}
}
