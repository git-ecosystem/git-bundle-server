package log

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Trace2 environment variables
const (
	// TODO: handle GIT_TRACE2 by adding a separate output config (see zapcore
	// "AdvancedConfiguration" example:
	// https://pkg.go.dev/go.uber.org/zap#example-package-AdvancedConfiguration)
	trace2Event string = "GIT_TRACE2_EVENT"
)

// Global start time
var globalStart = time.Now().UTC()

const trace2TimeFormat string = "2006-01-02T15:04:05.999999Z"

type ctxKey int

const (
	sidId ctxKey = iota
)

type Trace2 struct {
	logger *zap.Logger
}

func getTrace2OutputPaths(envKey string) []string {
	tr2Output := os.Getenv(envKey)

	// Configure the output
	if tr2, err := strconv.Atoi(tr2Output); err == nil {
		// Handle numeric values
		if tr2 == 1 {
			return []string{"stderr"}
		}
		// TODO: handle file handles 2-9 and unix sockets
	} else if tr2Output != "" {
		// Assume we received a path
		fileInfo, err := os.Stat(tr2Output)
		if err == nil && fileInfo.IsDir() {
			// If the path is an existing directory, generate a filename
			return []string{
				filepath.Join(tr2Output, fmt.Sprintf("trace2_%s.txt", globalStart.Format(trace2TimeFormat))),
			}
		} else {
			// Create leading directories
			parentDir := path.Dir(tr2Output)
			os.MkdirAll(parentDir, 0o755)
			return []string{tr2Output}
		}
	}

	return []string{}
}

func createTrace2ZapLogger() *zap.Logger {
	loggerConfig := zap.NewProductionConfig()

	// Configure the output for GIT_TRACE2_EVENT
	loggerConfig.OutputPaths = getTrace2OutputPaths(trace2Event)
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	// Encode UTC time
	loggerConfig.EncoderConfig.TimeKey = "time"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoder(
		func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(trace2TimeFormat))
		},
	)

	// Re-purpose the "message" to represent the (always-present) "event" key
	loggerConfig.EncoderConfig.MessageKey = "event"

	// Don't print the log level
	loggerConfig.EncoderConfig.LevelKey = ""

	// Disable caller info, we'll customize those fields manually
	logger, _ := loggerConfig.Build(zap.WithCaller(false))
	return logger
}

func NewTrace2() traceLoggerInternal {
	return &Trace2{
		logger: createTrace2ZapLogger(),
	}
}

type fieldList []zap.Field

func (l fieldList) withTime() fieldList {
	return append(l, zap.Float64("t_abs", time.Since(globalStart).Seconds()))
}

func (l fieldList) with(f ...zap.Field) fieldList {
	return append(l, f...)
}

func (t *Trace2) sharedFields(ctx context.Context) (context.Context, fieldList) {
	fields := fieldList{}

	// Get the session ID
	var sid uuid.UUID
	haveSid := false
	sidAny := ctx.Value(sidId)
	if sidAny != nil {
		sid, haveSid = sidAny.(uuid.UUID)
	}
	if !haveSid {
		sid = uuid.New()
		ctx = context.WithValue(ctx, sidId, sid)
	}
	fields = append(fields, zap.String("sid", sid.String()))

	// Hardcode the thread to "main" because Go doesn't like to share its
	// internal info about threading.
	fields = append(fields, zap.String("thread", "main"))

	// Get the caller of the function in trace2.go
	// Skip up two levels:
	// 0: this function
	// 1: the caller of this function (StartTrace, LogEvent, etc.)
	// 2: the function calling this trace2 library
	_, fileName, lineNum, ok := runtime.Caller(2)
	if ok {
		fields = append(fields,
			zap.String("file", filepath.Base(fileName)),
			zap.Int("line", lineNum),
		)
	}

	return ctx, fields
}

func (t *Trace2) logStart(ctx context.Context) context.Context {
	ctx, sharedFields := t.sharedFields(ctx)

	t.logger.Info("start", sharedFields.withTime().with(
		zap.Strings("argv", os.Args),
	)...)

	return ctx
}

func (t *Trace2) logExit(ctx context.Context, exitCode int) {
	_, sharedFields := t.sharedFields(ctx)
	fields := sharedFields.with(
		zap.Int("code", exitCode),
	)
	t.logger.Info("exit", fields.withTime()...)
	t.logger.Info("atexit", fields.withTime()...)

	t.logger.Sync()
}

func (t *Trace2) Exit(ctx context.Context, exitCode int) {
	t.logExit(ctx, exitCode)
	os.Exit(exitCode)
}

func (t *Trace2) Fatal(ctx context.Context, err error) {
	t.Exit(ctx, 1)
}

func (t *Trace2) Fatalf(ctx context.Context, format string, a ...any) {
	t.Exit(ctx, 1)
}
