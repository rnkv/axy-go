package axy

type Life interface {
	base() *Base
	OnSpawn()
	OnSpawned()
	// Do(function func()) chan bool
	OnMessage(message any)
	OnCancel()
	OnCanceled()
	OnDestroy()
	OnDestroyed()
}
