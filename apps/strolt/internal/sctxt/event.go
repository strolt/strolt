package sctxt

import "slices"

// EventType identifies a lifecycle event of an operation.
type EventType string //	@name	EventType

// Lifecycle events emitted during operation, source and destination phases.
const (
	EvOperationStart EventType = "OPERATION_START"
	EvOperationStop  EventType = "OPERATION_STOP"
	EvOperationError EventType = "OPERATION_ERROR"

	EvSourceStart EventType = "SOURCE_START"
	EvSourceStop  EventType = "SOURCE_STOP"
	EvSourceError EventType = "SOURCE_ERROR"

	EvDestinationStart EventType = "DESTINATION_START"
	EvDestinationStop  EventType = "DESTINATION_STOP"
	EvDestinationError EventType = "DESTINATION_ERROR"
)

// IsContextEventAvaliable reports whether the given event is a known context event.
func IsContextEventAvaliable(event EventType) bool {
	availableList := []EventType{
		EvOperationStart,
		EvOperationStop,
		EvOperationError,

		EvSourceStart,
		EvSourceStop,
		EvSourceError,

		EvDestinationStart,
		EvDestinationStop,
		EvDestinationError,
	}

	return slices.Contains(availableList, event)
}
