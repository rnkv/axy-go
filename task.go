package axy

// Task is an internal mailbox item representing a function scheduled via Base.Do.
type task struct {
	callable func()
	done     chan bool
}

func newTask(callable func()) task {
	return task{
		callable: callable,
		done:     make(chan bool, 1),
	}
}
