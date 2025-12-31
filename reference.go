package axy

type Reference interface {
	Key() string
	Cancel()
}
