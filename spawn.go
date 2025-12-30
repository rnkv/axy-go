package axy

func Spawn(life Life) Life {
	globalSystem.spawn(life, nil)
	return life
}
