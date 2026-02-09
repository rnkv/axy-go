## `parent_send`

Demonstrates a very common pattern:

- a **supervisor** actor spawns a **child** actor
- the child reports progress/events to its parent via `Parent().Send(...)`
- background goroutines use `Do(func(){ ... })` to hop back into the actor loop

### How to run

From the repository root:

```bash
go run ./examples/parent_send
```

Or from inside this folder:

```bash
make run
```

For contract checks (recommended while developing):

```bash
make run-debug
```

### Files

```text
examples/parent_send/
  main.go
  actors/
    counters/
      counter.go
      counter_messages.go
    loggers/
      logger.go
    supervisors/
      supervisor__counter.go
      supervisor__logger.go
      supervisor.go
```

### What to look for

- `actors/counters/counter_messages.go` — the counter’s protocol (`Start`, `Stop`) and outbound events (`Value`, `Destroyed`)
- `actors/counters/counter.go` — `Parent().Send(...)` performed from the actor goroutine (inside `Do(...)`)
- `actors/loggers/logger.go` — a tiny actor that receives events via `Reference.Send(...)`
- `actors/supervisors/supervisor.go` — parent spawns child in `OnSpawn()` and handles child events in `OnMessage`

