package axy

// Task is an internal mailbox item representing a function scheduled via Base.Do.
type Task struct {
	callable func()
	done     chan bool
}
