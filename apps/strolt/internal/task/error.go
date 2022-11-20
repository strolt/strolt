package task

import "fmt"

func errorAlreadyWorking(serviceName string, taskName string) error {
	return fmt.Errorf("task '%s' for service '%s' already working", taskName, serviceName)
}
