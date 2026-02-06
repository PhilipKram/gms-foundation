package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ConfigSchema defines the logger configuration options.
type ConfigSchema struct {
	// Level sets the minimum log level (-1=trace, 0=debug, 1=info, 2=warn, 3=error, 4=fatal, 5=panic)
	Level int8
	// Logstash enables Logstash-compatible JSON output format
	Logstash bool
	// Writer allows custom output writer (defaults to os.Stdout if nil)
	Writer io.Writer
	// DisableCaller disables automatic caller annotation
	DisableCaller bool
	// DisableTimestamp disables automatic timestamp annotation
	DisableTimestamp bool
}

// SetupLogger configures the global zerolog logger with the provided configuration.
// This function modifies the global log.Logger instance.
func SetupLogger(loggingConfig ConfigSchema) {
	zerolog.SetGlobalLevel(zerolog.Level(loggingConfig.Level))

	log.Logger = createBaseLogger(loggingConfig)
	if loggingConfig.Logstash {
		log.Logger = log.Logger.Hook(NewLevelValueHook())
	}
}

// New creates and returns a new logger instance with the provided configuration.
// Unlike SetupLogger, this does not modify the global logger.
func New(loggingConfig ConfigSchema) zerolog.Logger {
	logger := createBaseLogger(loggingConfig)
	if loggingConfig.Logstash {
		logger = logger.Hook(NewLevelValueHook())
	}
	return logger
}

func createBaseLogger(loggingConfig ConfigSchema) zerolog.Logger {
	var loggerWriter io.Writer

	// Use custom writer if provided, otherwise default to os.Stdout
	output := loggingConfig.Writer
	if output == nil {
		output = os.Stdout
	}

	if loggingConfig.Logstash {
		loggerWriter = output
	} else {
		loggerWriter = zerolog.ConsoleWriter{Out: output, TimeFormat: time.StampNano}
	}

	zerolog.TimeFieldFormat = time.RFC3339
	logsStructureUpdate()

	// Build logger with optional features
	ctx := zerolog.New(loggerWriter).With()

	if !loggingConfig.DisableTimestamp {
		ctx = ctx.Timestamp()
	}

	if !loggingConfig.DisableCaller {
		ctx = ctx.Caller()
	}

	return ctx.Logger()
}

func logsStructureUpdate() {
	zerolog.TimestampFieldName = "@timestamp"
	zerolog.LevelTraceValue = "TRACE"
	zerolog.LevelDebugValue = "DEBUG"
	zerolog.LevelInfoValue = "INFO"
	zerolog.LevelWarnValue = "WARN"
	zerolog.LevelErrorValue = "ERROR"
	zerolog.LevelFatalValue = "FATAL"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
}

// LevelValueHook adds numeric level values to log entries for Logstash/ELK compatibility.
// This hook adds a "level_value" field with numeric values that match Logstash conventions.
type LevelValueHook struct {
	levelValues map[zerolog.Level]int
}

// NewLevelValueHook creates a new LevelValueHook with standard Logstash level values.
// Level mappings: DEBUG=10000, INFO=20000, WARN=30000, ERROR=40000, FATAL=50000
func NewLevelValueHook() LevelValueHook {
	return LevelValueHook{
		levelValues: map[zerolog.Level]int{
			zerolog.DebugLevel: 10000,
			zerolog.InfoLevel:  20000,
			zerolog.WarnLevel:  30000,
			zerolog.ErrorLevel: 40000,
			zerolog.FatalLevel: 50000,
		},
	}
}

// Run implements the zerolog.Hook interface, adding the level_value field to each log entry.
func (h LevelValueHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if val, ok := h.levelValues[level]; ok {
		e.Int("level_value", val)
	}
}
