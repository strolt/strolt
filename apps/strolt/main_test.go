package main

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	// both ignores are for leaks which are detected locally
	goleak.VerifyTestMain(
		m,
		goleak.IgnoreTopFunction("github.com/umputun/remark42/backend/app.init.0.func1"),
		goleak.IgnoreTopFunction("net/http.(*Server).Shutdown"),
	)
}
