package rlog

// Interface that a replicated log library must provide.

import (
	"fmt"
)

type LogEntry = []byte
type Error = uint64

const (
	ENone = 0
	ENotPrimary = 1
)

type ReplicaServer interface {
	// returns ENotPrimary to tell the client to try a different server
	Append(e LogEntry) Error

	// returns the next log entry that this function has not previously returned
	// on this particular ReplicaServer.
	GetNextLogEntry() []LogEntry
}
