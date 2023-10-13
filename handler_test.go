package fingerscrossed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"testing"
)

func TestHandler(t *testing.T) {
	testcases := []struct {
		name       string
		ops        []loggerOp
		assertions []logsAssertion
	}{
		{
			name: "stay crossed",
			ops: []loggerOp{
				logAttrs(slog.LevelDebug, "test message debug"),
				logAttrs(slog.LevelInfo, "test message info"),
				logAttrs(slog.LevelWarn, "test message warn"),
			},
			assertions: []logsAssertion{
				logsAreEmpty(),
			},
		},
		{
			name: "uncross last",
			ops: []loggerOp{
				logAttrs(slog.LevelDebug, "test message debug"),
				logAttrs(slog.LevelInfo, "test message info"),
				logAttrs(slog.LevelWarn, "test message warn"),
				logAttrs(slog.LevelError, "test message error"),
			},
			assertions: []logsAssertion{
				assertByLog(
					[]logAssertion{logMessageEqual("test message debug")},
					[]logAssertion{logMessageEqual("test message info")},
					[]logAssertion{logMessageEqual("test message warn")},
					[]logAssertion{logMessageEqual("test message error")},
				),
			},
		},
		{
			name: "uncross first",
			ops: []loggerOp{
				logAttrs(slog.LevelError, "test message error"),
				logAttrs(slog.LevelDebug, "test message debug"),
				logAttrs(slog.LevelInfo, "test message info"),
				logAttrs(slog.LevelWarn, "test message warn"),
			},
			assertions: []logsAssertion{
				assertByLog(
					[]logAssertion{logMessageEqual("test message error")},
					[]logAssertion{logMessageEqual("test message debug")},
					[]logAssertion{logMessageEqual("test message info")},
					[]logAssertion{logMessageEqual("test message warn")},
				),
			},
		},
		{
			name: "uncross inside withArgs",
			ops: []loggerOp{
				logAttrs(slog.LevelDebug, "test message debug"),
				logAttrs(slog.LevelInfo, "test message info"),
				logAttrs(slog.LevelWarn, "test message warn"),
				withArgs([]any{slog.String("whatever", "test")}, logAttrs(slog.LevelError, "test message error")),
			},
			assertions: []logsAssertion{
				assertByLog(
					[]logAssertion{logMessageEqual("test message debug")},
					[]logAssertion{logMessageEqual("test message info")},
					[]logAssertion{logMessageEqual("test message warn")},
					[]logAssertion{logMessageEqual("test message error"), logHasAttr("whatever", "test")},
				),
			},
		},
		{
			name: "uncross inside withGroup",
			ops: []loggerOp{
				logAttrs(slog.LevelDebug, "test message debug"),
				logAttrs(slog.LevelInfo, "test message info"),
				logAttrs(slog.LevelWarn, "test message warn"),
				withGroup("testGroup", logAttrs(slog.LevelError, "test message error", slog.String("whatever", "test"))),
			},
			assertions: []logsAssertion{
				assertByLog(
					[]logAssertion{logMessageEqual("test message debug")},
					[]logAssertion{logMessageEqual("test message info")},
					[]logAssertion{logMessageEqual("test message warn")},
					[]logAssertion{logMessageEqual("test message error"), logHasAttrInGroup("testGroup", "whatever", "test")},
				),
			},
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			logBuffer := bytes.NewBuffer(nil)
			logger := slog.New(NewHandler(slog.NewJSONHandler(logBuffer, &slog.HandlerOptions{
				AddSource:   false,
				Level:       slog.LevelError,
				ReplaceAttr: nil,
			})))

			for _, op := range tt.ops {
				op(t, logger)
			}

			var logs []map[string]any
			decoder := json.NewDecoder(logBuffer)
			for {
				var log map[string]any
				err := decoder.Decode(&log)
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				logs = append(logs, log)
			}

			for _, assertion := range tt.assertions {
				assertion(t, logs)
			}
		})
	}
}

type loggerOp = func(t *testing.T, logger *slog.Logger)

func logAttrs(level slog.Level, msg string, attrs ...slog.Attr) loggerOp {
	return func(t *testing.T, logger *slog.Logger) {
		logger.LogAttrs(context.Background(), level, msg, attrs...)
	}
}

func withArgs(args []any, ops ...loggerOp) loggerOp {
	return func(t *testing.T, logger *slog.Logger) {
		l := logger.With(args...)
		for _, op := range ops {
			op(t, l)
		}
	}
}

func withGroup(name string, ops ...loggerOp) loggerOp {
	return func(t *testing.T, logger *slog.Logger) {
		l := logger.WithGroup(name)
		for _, op := range ops {
			op(t, l)
		}
	}
}

type logsAssertion = func(t *testing.T, logs []map[string]any)

func logsAreEmpty() logsAssertion {
	return func(t *testing.T, logs []map[string]any) {
		require.Empty(t, logs)
	}
}

func assertByLog(assertionsByLog ...[]logAssertion) logsAssertion {
	return func(t *testing.T, logs []map[string]any) {
		require.Len(t, logs, len(assertionsByLog))
		for i := range logs {
			log := logs[i]
			for _, assertion := range assertionsByLog[i] {
				assertion(t, log)
			}
		}
	}
}

type logAssertion func(t *testing.T, log map[string]any)

func logMessageEqual(msg string) logAssertion {
	return func(t *testing.T, log map[string]any) {
		require.Equal(t, msg, log["msg"])
	}
}

func logHasAttr(key string, value any) logAssertion {
	return func(t *testing.T, log map[string]any) {
		require.Equal(t, value, log[key])
	}
}

func logHasAttrInGroup(groupName string, key string, value any) logAssertion {
	return func(t *testing.T, log map[string]any) {
		group := log[groupName].(map[string]any)
		require.Equal(t, value, group[key])
	}
}
