package fingerscrossed_test

import (
	fingerscrossed "github.com/hectorj/slog-fingerscrossed"
	"log/slog"
	"os"
)

func ExampleNewHandler() {
	baseHandler := slog.NewJSONHandler(os.Stderr, nil)

	fingerscrossedHandler := fingerscrossed.NewHandler(baseHandler)

	logger := slog.New(fingerscrossedHandler)

	doThingsWithLogger(logger)
}

func doThingsWithLogger(_ *slog.Logger) {}
