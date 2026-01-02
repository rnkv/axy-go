package axy

// Envelope is an internal mailbox item representing a delivered message.
//
// It carries the sender identity alongside the message payload.
type Envelope struct {
	sender  Reference
	message any
}
