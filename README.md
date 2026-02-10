# Axy

Axy is a small, pragmatic **actor-style runtime for Go** built around a simple rule:

- **one mailbox + one event-loop goroutine per actor**

It is designed for building concurrent components with **serialized state access** and **message-driven coordination**, without bringing a heavy framework.

The name **axy** comes from **“ac sy”** — shorthand for **“actor system”**.

## What you get

- **Sequential message processing per actor**: actor state is touched by one goroutine (the actor loop).
- **Typed messages** (by convention): use Go types / structs as messages.
- **Lifecycle hooks**: spawn, message delivery, cancel, destroy.
- **Graceful shutdown**:
  - background goroutines started via `Go()` are awaited
  - child actors are canceled as part of parent shutdown
- **System scoping**: use a global system (`axy.Spawn`) or create isolated systems (`axy.NewSystem`).
- **Optional contract checks** (build tag `debug`): detect “called from wrong goroutine” mistakes early.

## Install

```bash
go get github.com/rnkv/axy-go
```

Import:

```go
import "github.com/rnkv/axy-go"
```

## Quick start

Start with the runnable example in `examples/parent_send/`:

```bash
go run ./examples/parent_send
```

It demonstrates the intended style:

- actors embed `axy.Base`
- actors talk to each other by sending typed messages
- parent/child messaging is done via `Parent().Send(...)`
- if you need to do something from a background goroutine, you hop back into the actor loop via `Do(func(){ ... })`

Key pieces (see the full files in `examples/parent_send/`):

```go
// examples/parent_send/actors/counters/counter_messages.go
package counters

import (
	"time"
)

type Start struct{ Every time.Duration }
type Stop struct{}

type Value struct{ N int } // child -> parent event
type Destroyed struct{}
```

## Recommended project structure (important)

Real-world actor systems get confusing fast if you don’t enforce a consistent layout. A pattern that works well (and is used in production codebases) is:

- group actors by **domain** (package/folder)
- keep each actor’s **message protocol** in a dedicated `{actor}_messages.go`
- store “how I talk to other actors” as explicit fields (raw references or typed handles)

Example:

```text
internal/
  actors/
    supervisors/
      supervisor.go
      supervisor__counter.go
      supervisor__logger.go
    counters/
      counter.go
      counter_messages.go
    loggers/
      logger.go
      logger_messages.go
```

Naming convention:

- actor packages are **plural** (`counters`, `supervisors`, `loggers`, `workers`, `clients`, `servers`, …)
  - it reads as “a set of actors of this kind”
  - it also reads like an **actors’ package**: the “home” where this kind of actor lives
  - it avoids awkward stutter like `worker.Worker`

If you want a concrete, runnable reference, see `examples/parent_send/` which mirrors this layout under `examples/parent_send/actors/...`.

Suggested file responsibilities inside one actor package:

- `{actor}_messages.go`
  - inbound command/query messages (what this actor accepts)
  - outbound events/responses (what this actor emits)
  - keep them small and stable — it’s the actor’s public protocol
- `{actor}.go` (actor implementation)
  - actor struct (embedding `axy.Base`)
  - `OnSpawn` (initialization, spawning children, wiring references, starting workers)
  - `OnMessage` (protocol handling)
  - helpers/private methods
- `{actor}__{peer}.go` (optional)
  - treat it as a **first-class domain entity** that represents *this actor’s relationship* to another actor
  - typically wraps an `axy.Reference` and provides intention-revealing **private helpers** (e.g. `startCounter`, `stopCounter`, `sendWork`, `retryWithBackoff`) for the actor implementation
  - it may also contain **supervisor-specific policy/logic** around that peer (timeouts, retries, backpressure rules, routing, metrics)

For example, in `supervisors/` you can keep a dedicated wrapper for the counter in a file like `supervisor__counter.go` and put all “how the supervisor talks to counters” logic there.

### Avoiding import cycles

In Go it’s easy to create cycles if two actor packages import each other’s message types. A practical rule:

- the **parent** imports the **child** package (to send commands / handle child events)
- the **child should not import the parent package**

If siblings need to exchange messages, put shared protocol types into a third package (e.g. `internal/protocol/...`) and have both import that.

## How actors “see” other actors (internal representations)

You have a few common options. Pick one and be consistent.

### Option A: store raw `axy.Reference`

This is the simplest and mirrors many production codebases:

```go
type Supervisor struct {
	axy.Base
	validator axy.Reference
	worker    axy.Reference
}
```

You then call `ref.Send(SomeMessage{...}, s)` from inside the actor goroutine.

### Option B: use `Parent().Send(...)` for child → parent events

If an actor is spawned as a child and needs to report progress/events upward, use `Parent().Send(...)` **from the actor goroutine** (for example, inside `OnMessage` or inside a `Do(...)` hop):

```go
func (c *Child) OnMessage(msg any, sender axy.Reference) {
	_ = sender
	_ = c.Parent().Send(ChildEvent{...})
}
```

This keeps the “child reports to parent” path explicit and avoids passing references down manually.

### Option C: `{actor}__{peer}.go` peer entity (recommended for readability)

If raw `Send` calls start to spread everywhere, encapsulate “how this actor talks to that actor”
into a **peer entity** stored in `{actor}__{peer}.go`.

This is the pattern used in `examples/parent_send/actors/supervisors/`:

```go
// supervisors/supervisor__counter.go
type counter struct {
	axy.Reference
	supervisor *Supervisor
}

func newCounter(supervisor *Supervisor) *counter {
	return &counter{
		Reference:   supervisor.Spawn(counters.NewCounter()),
		supervisor: supervisor,
	}
}

func (c *counter) start(every time.Duration) bool {
	return c.Send(counters.Start{Every: every}, c.supervisor)
}

func (c *counter) stop() bool {
	return c.Send(counters.Stop{}, c.supervisor)
}
```

The methods are usually **private** and encode your local policy (timeouts, retries, routing, etc.).

## Core concepts

### Actor

An `axy.Actor` is the behavior contract. Typical usage is:

- define your actor as a struct embedding `axy.Base`
- override the hooks you need (usually `OnMessage`, optionally spawn/cancel/destroy hooks)
- spawn it via `axy.Spawn` or `System.Spawn`
- interact with it via `axy.Reference`

See `actor.go` for the interface.

### Base (embeddable runtime)

`axy.Base` is the runtime implementation you embed into your actor. It provides:

- **mailbox** (buffered channel) and **event loop**
- `Send(message, sender)` to enqueue messages
- `Do(func())` to run a function **on the actor goroutine**
- `Go(func())` to start an actor-managed background goroutine (awaited on shutdown)
- `Cancel()` to request shutdown
- `Spawn(actor)` to spawn a **child actor** (from inside the actor goroutine)
- `Ctx()` context canceled during shutdown (use it to stop background work)

### Reference (public handle)

`axy.Reference` is a safe-to-share handle to an actor. It exposes:

- `Send(message, sender) bool`
- `Cancel()`
- `Key() string`

In the actor model used by axy, **messages are meant to be sent by actors**. A `Reference` is the handle that actors use to talk to other actors.

If you’re in a background goroutine (started with `Go()` or any other goroutine), use `Do(func(){ ... })` to hop back to the actor loop and send from there.

### System vs global helpers

There is a **global system** used by package-level helpers:

- `axy.Spawn(actor) Reference`
- `axy.Wait()`
- `axy.Done() <-chan struct{}`

If you want isolation/scoping (e.g. tests, multiple independent groups), create a system:

```go
s := axy.NewSystem()
ref := s.Spawn(NewMyActor())
ref.Cancel()
s.Wait()
```

### Parent handle (child → parent)

Inside a child actor you can get a handle to its parent with:

- `p := b.Parent()` (must be called on the actor goroutine)
- `p.Send(msg)` (sender is automatically the child actor)

This is useful for **reporting progress / events** from child to parent without
passing the parent reference down explicitly.

## Lifecycle hooks and shutdown semantics

### Lifecycle hooks (in order)

All hooks run on the **actor goroutine**:

1. `OnSpawn()` — beginning of actor goroutine, before loop
2. `OnSpawned()` — right after `OnSpawn()`
3. `OnMessage(message, sender)` — for each delivered message
4. `OnCancel()` — once when cancellation is requested
5. `OnCanceled()` — right after `OnCancel()`
6. `OnDestroy()` — once, when actor loop is about to exit
7. `OnDestroyed()` — right after `OnDestroy()`

Practical guideline:

- Put **spawn/initialization logic** (e.g. `Spawn(...)` child actors, wire references, start `Go()` workers) in `OnSpawn()`. It runs at the earliest point on the actor goroutine, before the message loop begins.

### What happens on `Cancel()`

Cancellation is cooperative and designed for graceful shutdown:

- The actor is marked as canceled; `OnCancel()` and `OnCanceled()` are called.
- `Ctx()` is canceled so your background work can stop.
- The runtime waits for:
  - **child actors** to be destroyed
  - **tracked goroutines** started via `Go()` to finish
- Then the actor loop is stopped and destroy hooks are executed.

Important detail: `Go()` becomes a no-op after shutdown reaches the “no new work” phase.

### Child actors

- A child actor is spawned from a parent actor via `Base.Spawn(child)`.
- Child actors are automatically tied to the parent shutdown.
- The parent does not fully destroy until all children are destroyed.

## Concurrency rules (and how to not shoot yourself)

### “Inside” vs “outside” the actor goroutine

Some operations are only valid from certain contexts:

- **From outside** (any goroutine):
  - `Cancel()`
  - `Do(func())` (by design: it hops to the actor loop)
- **From inside** (the actor goroutine):
  - `Send(...)` (actor-to-actor messaging)
  - `Spawn(child)`
  - `Parent()`

In other words: if you want to send a message “from a goroutine”, don’t — send it **from an actor**, and use `Do()` as the bridge.

Contract note about `Do()`:

- **Do not call `Do()` from the actor goroutine.** `Do()` is an “outside → inside” hop.
- Calling it from inside can lead to mailbox pressure (it enqueues into the same mailbox you are currently draining), and it’s easy to accidentally deadlock by waiting on the returned channel while the actor loop can’t progress.

### Debug build tag: contract checks

If you build with the `debug` tag, axy will panic on common contract violations:

- calling “inside-only” APIs from outside the actor goroutine
- calling “outside-only” APIs from inside the actor goroutine

Under the hood (debug builds only), axy tracks the actor loop goroutine and compares it to the current goroutine using a goroutine ID extracted from the runtime stack trace. This makes “wrong goroutine” mistakes fail fast during development.

Example:

```bash
# Run an application / example with contract checks enabled:
go run -tags=debug ./examples/parent_send

# Or run tests with the same checks:
go test -tags=debug ./...
```

You can (and usually should) combine this with the race detector:

```bash
go run -race -tags=debug ./examples/parent_send
go test -race -tags=debug ./...
```

Without `debug`, these assertions are no-ops.

## Mailbox behavior and ordering

- Each actor has a **buffered mailbox** (currently `128` items).
- `Send` and `Do` enqueue into that mailbox.
- Processing is **sequential**: one item at a time, in the order items are taken from the mailbox.

Notes:

- **`Send` may block** if the mailbox buffer is full (it is a channel send).
- Ordering across multiple concurrent senders is not guaranteed beyond Go channel semantics (i.e. whichever send happens first wins).
- `Send(nil, ...)` is rejected and returns `false`.
- `Send(...)` returns `false` if the actor is already canceled/shutting down.

## Patterns

### Serialized state mutation with `Do`

If you need to mutate actor state (or send messages) from a background goroutine, hop into the actor loop:

```go
done := actorBase.Do(func() {
	// safely touches actor state / sends messages on the actor goroutine
})
ok := <-done
_ = ok // false means actor is shutting down and the task did not run
```

Reminder: `Do()` is intended to be called **from outside** the actor goroutine. If you are already inside `OnMessage` (or any actor hook), just run the code directly.

### Background work tied to actor lifetime

Use `Go()` and `Ctx()`:

```go
a.Go(func() {
	for {
		select {
		case <-a.Ctx().Done():
			return
		default:
			// do work
		}
	}
})
```

## Logging

`Base`’s default hook implementations log lifecycle events via `log/slog`.

To override the logger:

```go
axy.SetLogger(myLogger)
```

If you override hooks like `OnMessage`, you can log however you want (and still call `Base.OnMessage` if you want the default log).

## Pitfalls / gotchas

- **`SetKey` is optional**: it’s primarily for identifying different instances of the same actor type (logging/tracing/diagnostics).
- **If you use `SetKey`, call it before spawning**: calling it more than once panics.
- **Don’t block the actor loop**: `OnMessage` should be quick; offload slow work to `Go()` and communicate back with messages.
- **Beware mailbox backpressure**: `Send` can block when the mailbox is full.
- **Use `System` for isolation**: tests and multiple independent groups should prefer `NewSystem()`.
- **Use `-tags=debug` while developing**: it catches “wrong goroutine” usage early.

## Waiting for shutdown

- Global system:
  - `axy.Wait()` blocks until all actors are destroyed
  - `axy.Done()` returns a channel closed when shutdown completes
- Custom system:
  - `s.Wait()` / `s.Done()`

