package axy

func Spawn(life Life) Life {
	return globalSystem.spawn(life, nil)
}
