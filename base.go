package axy

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

type Base struct {
	system                    *System
	parent                    *Base
	life                      Life
	spawnOnce                 sync.Once
	queue                     chan any
	initializeQueueOnce       sync.Once
	externalCtx               context.Context
	externalCancel            context.CancelFunc
	initializeExternalCtxOnce sync.Once
	childrenCtx               context.Context
	childrenCancel            context.CancelFunc
	initializeChildrenCtxOnce sync.Once
	childrenWG                sync.WaitGroup
	childrenWGMutex           sync.Mutex
	isChildrenWGLocked        bool
	internalCtx               context.Context
	internalCancel            context.CancelFunc
	initializeInternalCtxOnce sync.Once
	isCancelRequested         atomic.Bool
}

func (b *Base) base() *Base {
	return b
}

func (b *Base) name() string {
	if b.life == nil {
		return "UnknownActor"
	}

	t := reflect.TypeOf(b.life)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

func (b *Base) initializeExternalCtx() {
	b.initializeExternalCtxOnce.Do(func() {
		var parentChildrenCtx context.Context

		if b.parent == nil {
			parentChildrenCtx = context.Background()
		} else {
			parentChildrenCtx = b.parent.childrenCtx
		}

		b.externalCtx, b.externalCancel = context.WithCancel(parentChildrenCtx)
	})
}

func (b *Base) initializeChildrenCtx() {
	b.initializeChildrenCtxOnce.Do(func() {
		b.childrenCtx, b.childrenCancel = context.WithCancel(context.Background())
	})
}

func (b *Base) initializeInternalCtx() {
	b.initializeInternalCtxOnce.Do(func() {
		b.internalCtx, b.internalCancel = context.WithCancel(context.Background())
	})
}

func (b *Base) initializeQueue() {
	b.initializeQueueOnce.Do(func() {
		b.queue = make(chan any, 128)
	})
}

func (b *Base) OnSpawn() {
	logger.Debug(fmt.Sprintf("%s spawning...", b.name()))
}

func (b *Base) OnSpawned() {
	logger.Debug(fmt.Sprintf("%s spawned.", b.name()))
}

func (b *Base) live() {
	b.life.OnSpawn()
	b.life.OnSpawned()
	b.loop()
	b.cleanUpQueue()
	b.life.OnDestroy()
	b.life.OnDestroyed()

	if b.parent != nil {
		b.parent.childrenWG.Done()
	}

	b.system.removeActor()
}

func (b *Base) cancel() {
	b.life.OnCancel()
	b.life.OnCanceled()
	b.childrenWGMutex.Lock()
	b.isChildrenWGLocked = true
	b.childrenWGMutex.Unlock()
	b.childrenCancel()

	go func() {
		b.childrenWG.Wait()
		b.internalCancel()
		b.queue <- nil
	}()
}

func (b *Base) handle(object any) bool {
	switch object := object.(type) {
	case nil:
		return false
	case Task:
		object.callable()
		object.done <- true
		return true
	default:
		b.life.OnMessage(object)
		return true
	}
}

func (b *Base) loop() {
	externalCtxDone := b.externalCtx.Done()

	if b.externalCtx.Err() != nil {
		externalCtxDone = nil
		b.cancel()
	}

	for {
		select {
		case <-externalCtxDone:
			externalCtxDone = nil
			b.cancel()
		case object := <-b.queue:
			if !b.handle(object) {
				return
			}
		}
	}
}

func (b *Base) cleanUpQueue() {
	for {
		select {
		case object := <-b.queue:
			b.handle(object)
		default:
			return
		}
	}
}

func (b *Base) OnMessage(message any) {
	logger.Debug(fmt.Sprintf("%s received message. [%v]", b.name(), message))
}

func (b *Base) Do(callable func()) chan bool {
	b.initializeInternalCtx()
	b.initializeQueue()

	task := Task{
		callable: callable,
		done:     make(chan bool, 1),
	}

	if b.internalCtx.Err() != nil {
		task.done <- false
		return task.done
	}

	select {
	case <-b.internalCtx.Done():
		task.done <- false
		return task.done
	case b.queue <- task:
		return task.done
	}
}

func (b *Base) Reference() Reference {
	b.initializeExternalCtx()
	b.initializeQueue()

	return Reference{
		ctx:    b.externalCtx,
		cancel: b.externalCancel,
		queue:  b.queue,
	}
}

func (b *Base) Parent() Parent {
	if b.parent == nil {
		panic("Actor has no parent.")
	}

	b.initializeInternalCtx()
	b.initializeQueue()

	return Parent{
		ctx:   b.parent.internalCtx,
		queue: b.parent.queue,
	}
}

func (b *Base) Ctx() context.Context {
	b.initializeChildrenCtx()
	return b.childrenCtx
}

func (b *Base) Cancel() {
	b.externalCancel()
}

func (b *Base) OnCancel() {
	logger.Debug(fmt.Sprintf("%s canceling...", b.name()))
}

func (b *Base) OnCanceled() {
	logger.Debug(fmt.Sprintf("%s canceled.", b.name()))
}

func (b *Base) Go(callable func()) {
	b.childrenWGMutex.Lock()

	if b.isChildrenWGLocked {
		b.childrenWGMutex.Unlock()
		return
	}

	b.childrenWG.Add(1)
	b.childrenWGMutex.Unlock()

	go func() {
		defer b.childrenWG.Done()
		callable()
	}()
}

func (b *Base) OnDestroy() {
	logger.Debug(fmt.Sprintf("%s destroying...", b.name()))
}

func (b *Base) OnDestroyed() {
	logger.Debug(fmt.Sprintf("%s destroyed.", b.name()))
}

func (b *Base) Spawn(life Life) Reference {
	b.system.spawn(life, b)
	return life.base().Reference()
}
