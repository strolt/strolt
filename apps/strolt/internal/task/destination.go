package task

// IsAvailableDestinationName reports whether the destination is configured for the task.
func (t *Task) IsAvailableDestinationName(destinationName string) bool {
	for name := range t.TaskConfig.Destinations {
		if name == destinationName {
			return true
		}
	}

	return false
}
