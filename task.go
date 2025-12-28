package axy

type Task struct {
	callable func()
	done     chan bool
}
