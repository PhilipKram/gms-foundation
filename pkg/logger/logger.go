package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConfigSchema struct {
	Level    int8
	Logstash bool
}

func SetupLogger(loggingConfig ConfigSchema) {
	zerolog.SetGlobalLevel(zerolog.Level(loggingConfig.Level))

	log.Logger = createBaseLogger(loggingConfig)
	if loggingConfig.Logstash {
		log.Logger = log.Logger.Hook(NewLevelValueHook())
	}
}

func createBaseLogger(loggingConfig ConfigSchema) zerolog.Logger {
	var loggerWriter io.Writer
	if loggingConfig.Logstash {
		loggerWriter = os.Stdout
	} else {
		loggerWriter = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampNano}
	}

	zerolog.TimeFieldFormat = time.RFC3339
	logsStructureUpdate()

	return zerolog.New(loggerWriter).
		With().
		Timestamp().
		Caller().
		Logger()
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

type LevelValueHook struct {
	levelValues map[zerolog.Level]int
}

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

func (h LevelValueHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if val, ok := h.levelValues[level]; ok {
		e.Int("level_value", val)
	}
}
