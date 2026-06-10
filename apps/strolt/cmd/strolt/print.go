package cmd

import "fmt"

// Print writes the given value to stdout followed by a newline.
func Print(arg any) {
	fmt.Println(arg) //nolint:forbidigo
}

// Printf writes the formatted message to stdout followed by a newline.
func Printf(format string, args ...any) {
	fmt.Println(fmt.Sprintf(format, args...)) //nolint:forbidigo
}
