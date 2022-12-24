package cmd

import "fmt"

func Print(arg interface{}) {
	fmt.Println(arg) //nolint:forbidigo
}

func Printf(format string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(format, args...)) //nolint:forbidigo
}
