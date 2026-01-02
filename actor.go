package axy

type Actor interface {
	base() *Base
	Key() string
	Send(message any, sender Reference) bool
	Perception(perceiver Actor) Perception
	Cancel()
	OnSpawn()
	OnSpawned()
	// Do(function func()) chan bool
	OnMessage(message any, sender Reference)
	OnCancel()
	OnCanceled()
	OnDestroy()
	OnDestroyed()
}
