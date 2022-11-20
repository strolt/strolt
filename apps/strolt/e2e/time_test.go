package e2e_test

import (
	"log"
	"time"
)

type TimeTook struct {
	Message string
	Start   time.Time
}

func timeTook(message string) *TimeTook {
	return &TimeTook{
		Message: message,
		Start:   time.Now(),
	}
}

func (t *TimeTook) stop() {
	elapsed := time.Since(t.Start)
	log.Printf("%s %s", t.Message, elapsed)
}
