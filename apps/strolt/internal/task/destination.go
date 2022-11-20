package task

func (t Task) IsAvailableDestinationName(destinationName string) bool {
	for name := range t.TaskConfig.Destinations {
		if name == destinationName {
			return true
		}
	}

	return false
}
