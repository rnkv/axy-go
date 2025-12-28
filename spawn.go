package axy

func Spawn(life Life) Reference {
	globalSystem.spawn(life, nil)
	return life.base().Reference()
}
