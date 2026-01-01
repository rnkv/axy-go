package axy

func Spawn(actor Actor) Reference {
	return globalSystem.spawn(actor, nil)
}
