package axy

func Spawn(life Life) Reference {
	return globalSystem.spawn(life, nil)
}
