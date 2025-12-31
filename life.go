package axy

type Life interface {
	base() *Base
	Key() string
	Send(message any, sender Reference) bool
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
