package util

import (
	"time"

	"github.com/briandowns/spinner"
)

func NewSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithSuffix(" Loading...")) //nolint:gomnd
}
