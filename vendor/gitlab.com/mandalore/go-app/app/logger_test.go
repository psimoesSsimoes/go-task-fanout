package app

import (
	"fmt"
	"testing"
)

func TestNewLogger(t *testing.T) {
	newLogger("test", "ignore", "this")
	newLogger("test")
}

func TestDefaultLogger(t *testing.T) {
	// Logger.Reconfigure(log.WithTags("test", "tag"))
	Logger.LogCallers("this", 3)
	Logger.LogError(fmt.Errorf("some error"), "")
	Logger.LogError(
		NewError(
			ErrorDevPoo,
			"application error",
			NewError(
				ErrorUnexpected,
				"nested application error",
				fmt.Errorf("nested^2 error"),
			),
		),
		"a message",
	)

	t.Log("OK")
}
