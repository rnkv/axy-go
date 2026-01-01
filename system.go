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

func (s *System) Spawn(actor Actor) Reference {
	return s.spawn(actor, nil)
}

func (s *System) spawn(actor Actor, parent *Base) Reference {
	base := actor.base()

	base.spawnOnce.Do(func() {
		base.system = s
		base.parent = parent
		base.actor = actor
		s.addActor()

		if base.parent != nil {
			base.parent.childrenWG.Add(1)
		}

		base.initializeQueue()
		base.initializeExternalCtx()
		base.initializeChildrenCtx()
		base.initializeInternalCtx()
		go base.live()
	})

	return actor
}
