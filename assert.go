//go:build !debug

package axy

func goid() uint64 {
	return 0
}

// AssertInside panics if called outside the actor goroutine (debug only).
func (b *Base) AssertInside() {}

// AssertOutside panics if called inside the actor goroutine (debug only).
func (b *Base) AssertOutside() {}
