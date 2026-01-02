package axy

// Envelope is an internal mailbox item representing a delivered message.
//
// It carries the sender identity alongside the message payload.
type envelope struct {
	sender  Reference
	message any
}

func newEnvelope(sender Reference, message any) envelope {
	return envelope{
		sender:  sender,
		message: message,
	}
}
