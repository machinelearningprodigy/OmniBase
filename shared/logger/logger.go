package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new structured zap logger.
// In production it uses JSON encoding; in development it uses a human-readable console format.
func New(level, env string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.Set(level); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	if env == "production" {
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapLevel)
		return cfg.Build()
	}

	// Development: colored, human-readable output
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// MustNew creates a logger or panics. Used in main functions.
func MustNew(level, env string) *zap.Logger {
	l, err := New(level, env)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	return l
}
