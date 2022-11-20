package sctxt

type EventType string

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

	for _, e := range availableList {
		if e == event {
			return true
		}
	}

	return false
}
