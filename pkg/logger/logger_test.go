package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
)

func TestSetupLogger_SetsLevel(t *testing.T) {
	SetupLogger(Config{
		Level:            int8(zerolog.WarnLevel),
		Logstash:         true,
		Writer:           &bytes.Buffer{},
		DisableCaller:    true,
		DisableTimestamp: true,
	})

	if zerolog.GlobalLevel() != zerolog.WarnLevel {
		t.Errorf("expected global level Warn, got %v", zerolog.GlobalLevel())
	}

	// Reset global level so other tests are not affected
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}

func TestNew_ReturnsIndependentLogger(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	var buf1, buf2 bytes.Buffer

	l1 := New(Config{Level: int8(zerolog.InfoLevel), Logstash: true, Writer: &buf1, DisableCaller: true, DisableTimestamp: true})
	l2 := New(Config{Level: int8(zerolog.InfoLevel), Logstash: true, Writer: &buf2, DisableCaller: true, DisableTimestamp: true})

	l1.Info().Msg("from l1")
	l2.Info().Msg("from l2")

	if buf1.Len() == 0 {
		t.Error("expected l1 to write output")
	}
	if buf2.Len() == 0 {
		t.Error("expected l2 to write output")
	}
	if buf1.String() == buf2.String() {
		t.Error("expected different messages in each buffer")
	}
}

func TestLevelValueHook_AllLevels(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	hook := NewLevelValueHook()

	tests := []struct {
		level zerolog.Level
		want  int
	}{
		{zerolog.TraceLevel, 5000},
		{zerolog.DebugLevel, 10000},
		{zerolog.InfoLevel, 20000},
		{zerolog.WarnLevel, 30000},
		{zerolog.ErrorLevel, 40000},
		{zerolog.FatalLevel, 50000},
		{zerolog.PanicLevel, 60000},
	}

	for _, tt := range tests {
		// Verify the hook mapping directly
		if val, ok := hook.levelValues[tt.level]; !ok || val != tt.want {
			t.Errorf("level %v: expected value %d, got %d (ok=%v)", tt.level, tt.want, val, ok)
			continue
		}

		// For safe levels, verify the hook actually writes level_value
		if tt.level >= zerolog.DebugLevel && tt.level <= zerolog.ErrorLevel {
			var buf bytes.Buffer
			logger := zerolog.New(&buf).Hook(hook)
			logger.WithLevel(tt.level).Msg("test")

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Errorf("level %v: unmarshal error: %v (output: %q)", tt.level, err, buf.String())
				continue
			}
			if int(entry["level_value"].(float64)) != tt.want {
				t.Errorf("level %v: expected level_value %d, got %v", tt.level, tt.want, entry["level_value"])
			}
		}
	}
}

func TestNew_DisableCaller(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	var buf bytes.Buffer
	l := New(Config{Level: int8(zerolog.InfoLevel), Logstash: true, Writer: &buf, DisableCaller: true, DisableTimestamp: true})
	l.Info().Msg("no caller")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v (output: %q)", err, buf.String())
	}
	if _, ok := entry["caller"]; ok {
		t.Error("expected no caller field when DisableCaller=true")
	}
}

func TestNew_DisableTimestamp(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	var buf bytes.Buffer
	l := New(Config{Level: int8(zerolog.InfoLevel), Logstash: true, Writer: &buf, DisableCaller: true, DisableTimestamp: true})
	l.Info().Msg("no timestamp")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v (output: %q)", err, buf.String())
	}
	if _, ok := entry["@timestamp"]; ok {
		t.Error("expected no @timestamp field when DisableTimestamp=true")
	}
}

func TestNew_CustomWriter(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	var buf bytes.Buffer
	l := New(Config{Level: int8(zerolog.InfoLevel), Logstash: true, Writer: &buf, DisableCaller: true, DisableTimestamp: true})
	l.Info().Msg("custom writer test")

	if buf.Len() == 0 {
		t.Error("expected output in custom writer")
	}
}
