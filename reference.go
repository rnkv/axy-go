package axy

type Reference interface {
	Key() string
	Send(message any, sender Reference)
	Cancel()
}
