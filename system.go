package axy

import "sync"

type System struct {
	actorsCount      int
	actorsCountMutex sync.Mutex
	actorsCountCond  *sync.Cond
}

func NewSystem() *System {
	s := &System{}
	s.actorsCountCond = sync.NewCond(&s.actorsCountMutex)
	return s
}

var globalSystem = NewSystem()

func (s *System) addActor() {
	s.actorsCountMutex.Lock()
	s.actorsCount++
	s.actorsCountMutex.Unlock()
}

func (s *System) removeActor() {
	s.actorsCountMutex.Lock()
	s.actorsCount--

	if s.actorsCount == 0 {
		s.actorsCountCond.Broadcast()
	}

	s.actorsCountMutex.Unlock()
}

func (s *System) Wait() {
	s.actorsCountMutex.Lock()
	for s.actorsCount > 0 {
		s.actorsCountCond.Wait()
	}
	s.actorsCountMutex.Unlock()
}

func (s *System) Spawn(life Life) {
	s.spawn(life, nil)
}

func (s *System) spawn(life Life, parent *Base) {
	actor := life.base()

	actor.spawnOnce.Do(func() {
		// if actor.isSpawning || actor.isSpawned {
		// 	return
		// }

		// actor.isSpawning = true
		actor.system = s
		actor.parent = parent
		actor.life = life
		s.addActor()

		if actor.parent != nil {
			actor.parent.childrenWG.Add(1)
		}

		actor.initializeQueue()
		actor.initializeExternalCtx()
		actor.initializeChildrenCtx()
		actor.initializeInternalCtx()
		go actor.live()
	})
}
