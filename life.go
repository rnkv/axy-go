package axy

type Life interface {
	base() *Base
	Key() string
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
