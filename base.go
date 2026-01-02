package axy

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// Base is the embeddable implementation of the axy actor runtime.
//
// Embed Base into your actor struct to get default implementations of:
//   - message delivery via Send/OnMessage
//   - cancellation via Cancel/OnCancel/OnDestroy
//   - actor-scoped helpers (Ctx/Do/Go/Spawn)
//
// The runtime initializes Base when the actor is spawned.
type Base struct {
	key                       string
	system                    *System
	parent                    *Base
	actor                     Actor
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

// base returns the embedded Base.
//
// It exists to let the runtime access internal state even when Base is embedded
// anonymously into user-defined actor structs.
func (b *Base) base() *Base {
	return b
}

func (b *Base) kind() string {
	if b.actor == nil {
		return "Unknown"
	}

	t := reflect.TypeOf(b.actor)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

// SetKey sets a stable identifier for the actor.
//
// SetKey must be called before spawning (typically in your constructor).
// Calling SetKey more than once panics.
func (b *Base) SetKey(key string) {
	if b.key != "" {
		panic("Key already set.")
	}

	b.key = key
}

// Key returns the actor key previously set via SetKey.
//
// The key is optional but recommended for logging/tracing. Keys must be set at
// most once; calling SetKey twice panics.
func (b *Base) Key() string {
	return b.key
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

// OnSpawn is a lifecycle hook called at the beginning of the actor goroutine, before the loop starts.
func (b *Base) OnSpawn() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s spawning...", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s spawning...", b.kind()))
	}
}

// OnSpawned is a lifecycle hook called after OnSpawn, still on the actor goroutine.
func (b *Base) OnSpawned() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s spawned.", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s spawned.", b.kind()))
	}
}

func (b *Base) live() {
	b.actor.OnSpawn()
	b.actor.OnSpawned()
	b.loop()
	b.cleanUpQueue()
	b.actor.OnDestroy()
	b.actor.OnDestroyed()

	if b.parent != nil {
		b.parent.childrenWG.Done()
	}

	b.system.removeActor()
}

func (b *Base) cancel() {
	b.actor.OnCancel()
	b.actor.OnCanceled()
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
	case task:
		object.callable()
		object.done <- true
		return true
	case envelope:
		b.actor.OnMessage(object.message, object.sender)
		return true
	default:
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

// OnMessage is a lifecycle hook called for every delivered message, on the actor goroutine.
func (b *Base) OnMessage(message any, sender Reference) {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s received message.", b.kind()), "key", b.key, "message", message)
	} else {
		logger.Debug(fmt.Sprintf("%s received message.", b.kind()), "message", message)
	}
}

// Do schedules callable to be executed on the actor goroutine.
//
// This is useful for serialized access to actor state without sending a typed
// message. The returned channel receives:
//   - true if the task was executed
//   - false if the actor is already shutting down and the task could not run
func (b *Base) Do(callable func()) chan bool {
	b.initializeInternalCtx()
	b.initializeQueue()

	task := newTask(callable)

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

// Send enqueues a message to this actor.
//
// The sender is used for tracing/diagnostics and can be used by the receiver
// to reply. Returns false if message is nil or the actor is already canceled.
// func (b *Base) Send(message any, sender Reference) bool {
// 	if message == nil {
// 		return false
// 	}

// 	b.initializeExternalCtx()
// 	b.initializeQueue()

// 	if b.externalCtx.Err() != nil {
// 		return false
// 	}

// 	select {
// 	case <-b.externalCtx.Done():
// 		return false
// 	case b.queue <- newEnvelope(sender, message):
// 		return true
// 	}
// }

// Reference returns a handle to this actor.
//
// Returns a [Reference] handle that can be used to send messages to the actor.
func (b *Base) Reference() Reference {
	b.initializeExternalCtx()
	b.initializeQueue()

	return Reference{
		key:    b.key,
		ctx:    b.externalCtx,
		cancel: b.externalCancel,
		queue:  b.queue,
	}
}

// Parent returns a handle for the parent actor to send messages to this actor.
//
// Returns a [Parent] handle that can be used to send messages to the parent actor.
// Panics if the actor has no parent.
func (b *Base) Parent() Parent {
	if b.parent == nil {
		panic("Actor has no parent.")
	}

	b.parent.initializeInternalCtx()
	b.parent.initializeQueue()

	return Parent{
		child: b.actor.Reference(),
		ctx:   b.parent.internalCtx,
		queue: b.parent.queue,
	}
}

// Ctx returns a context that is canceled when the actor is shutting down.
//
// Use this context to stop background goroutines started via Go.
func (b *Base) Ctx() context.Context {
	b.initializeChildrenCtx()
	return b.childrenCtx
}

// Cancel requests actor shutdown.
//
// It is safe to call multiple times.
func (b *Base) Cancel() {
	b.externalCancel()
}

// OnCancel is a lifecycle hook called when cancellation is requested.
func (b *Base) OnCancel() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s canceling...", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s canceling...", b.kind()))
	}
}

// OnCanceled is a lifecycle hook called after OnCancel.
func (b *Base) OnCanceled() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s canceled.", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s canceled.", b.kind()))
	}
}

// Go runs callable in a new goroutine that is tracked by the actor.
//
// Tracked goroutines are awaited during shutdown, which lets OnCancel trigger
// graceful cleanup work. If shutdown has progressed to the point where new work
// should not start, Go becomes a no-op.
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

// OnDestroy is a lifecycle hook called when the actor is about to exit.
func (b *Base) OnDestroy() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s destroying...", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s destroying...", b.kind()))
	}
}

// OnDestroyed is a lifecycle hook called after OnDestroy.
func (b *Base) OnDestroyed() {
	if b.key != "" {
		logger.Debug(fmt.Sprintf("%s destroyed.", b.kind()), "key", b.key)
	} else {
		logger.Debug(fmt.Sprintf("%s destroyed.", b.kind()))
	}
}

// Spawn starts a child actor in the same system and returns its [Reference].
//
// The returned actor is automatically canceled if the parent actor is canceled.
func (b *Base) Spawn(actor Actor) Reference {
	return b.system.spawn(actor, b)
}
