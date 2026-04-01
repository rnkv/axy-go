package axy

func Initialize(actor Actor, key ...string) {
	actor.base().initializeActor(actor)

	if len(key) > 0 {
		actor.base().SetKey(key[0])
	}
}
