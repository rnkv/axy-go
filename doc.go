// Package axy provides a small actor-like runtime for building concurrent components
// that communicate by sending messages to each other.
//
// The core idea is:
//   - Implement your actor by embedding [Base] and overriding lifecycle hooks like
//     OnSpawn, OnMessage, OnCancel, and OnDestroy.
//   - Spawn actors via [Spawn] (global) or [System.Spawn] (custom system).
//   - Interact with spawned actors through the [Reference] interface.
//
// Concurrency model (high level):
//   - Each actor owns a single goroutine that processes its mailbox sequentially.
//   - Use Base.Send to enqueue a message for processing by the actor.
//   - Use Base.Do to enqueue a function to be executed on the actor goroutine.
//   - Use Base.Go for actor-managed background goroutines that are awaited during shutdown.
//
// Cancellation model (high level):
//   - Calling Cancel() requests shutdown.
//   - Base.Ctx() is canceled to signal background work to stop.
//   - After background work completes, the actor goroutine exits and the actor is destroyed.

package axy
