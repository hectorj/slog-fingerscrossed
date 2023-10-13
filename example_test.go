package fingerscrossed_test

import (
	fingerscrossed "github.com/hectorj/slog-fingerscrossed"
	"log/slog"
	"os"
)

func ExampleHandler() {
	// 2 scenarios: with and without error logs

	// Scenario 1: without error
	{
		baseHandler := slog.NewJSONHandler(os.Stderr, nil)
		fingerscrossedHandler := fingerscrossed.NewHandler(baseHandler)

		logger := slog.New(fingerscrossedHandler)

		logger.Debug("debug msg") // <-- no log output
		logger.Info("info msg")   // <-- no log output
		logger.Warn("warn msg")   // <-- no log output

		_ = fingerscrossedHandler.FlushLogs(slog.LevelInfo) // <-- outputs "info msg" and "warn msg" logs, but not "debug msg"
	}
	// Scenario 2: with error
	{
		baseHandler := slog.NewJSONHandler(os.Stderr, nil)
		fingerscrossedHandler := fingerscrossed.NewHandler(baseHandler)

		logger := slog.New(fingerscrossedHandler)

		logger.Debug("debug msg") // <-- no log output
		logger.Info("info msg")   // <-- no log output
		logger.Error("error msg") // <-- outputs "debug msg", "info msg", and "error msg" logs
		logger.Warn("warn msg")   // <-- outputs "warn msg" log

		_ = fingerscrossedHandler.FlushLogs(slog.LevelInfo) // <-- everything is already flushed, nothing happens
	}
}
