package axy

type Life interface {
	base() *Base
	Key() string
	OnSpawn()
	OnSpawned()
	// Do(function func()) chan bool
	OnMessage(sender Life, message any)
	OnCancel()
	OnCanceled()
	OnDestroy()
	OnDestroyed()
}
