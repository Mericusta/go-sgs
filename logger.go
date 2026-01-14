package sgs

import (
	"strings"

	"github.com/Mericusta/go-stp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	outputPaths []string
	fields      []zapcore.Field
	levelAt     zapcore.Level

	// zap 日志实例
	core *zap.Logger

	// zap 日志 hook
	hooks []func(*Logger)
}

var Log *Logger

func init() {
	Log = Log.New(
		Log.WithOutputPaths("stdout"),
		Log.WithFields(zap.String("identify", "sgs")),
	)
}

func (l *Logger) New(opts ...func(*Logger)) *Logger {
	// 构造实例
	logger := &Logger{}
	// 继承原实例
	if l != nil {
		logger.outputPaths = l.outputPaths
		logger.fields = l.fields
		logger.levelAt = l.levelAt
	}
	// 构造 core
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.OutputPaths = logger.outputPaths
	zapCfg.Level = zap.NewAtomicLevelAt(logger.levelAt)
	core, err := zapCfg.Build()
	if core == nil || err != nil {
		return nil
	}
	logger.core = core
	// 应用传进来的 option
	for _, opt := range opts {
		opt(logger)
	}
	return logger
}

func (*Logger) WithOutputPaths(paths ...string) func(*Logger) {
	return func(l *Logger) {
		// outputPaths 不能重复，不然会多次输出
		outputPathsArray := stp.NewArray(l.outputPaths)
		stp.NewArray(paths).ForEach(func(v string, i int) {
			if !outputPathsArray.Includes(v) {
				outputPathsArray.Push(v)
			}
		})
		l.outputPaths = outputPathsArray.Slice()
	}
}

func (*Logger) WithFields(fields ...zapcore.Field) func(*Logger) {
	return func(l *Logger) {
		for _, field := range fields {
			index := -1
			for _index, _field := range l.fields {
				if _field.Key == field.Key {
					index = _index
					break
				}
			}
			if index != -1 {
				l.fields[index] = field
			} else {
				l.fields = append(l.fields, field)
			}
		}
		l.core = l.core.With(l.fields...)
	}
}

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
)

func (*Logger) WithLevelAt(level string) func(*Logger) {
	return func(l *Logger) {
		switch strings.ToLower(strings.TrimSpace(level)) {
		case LevelDebug:
			l.levelAt = zap.DebugLevel
		case LevelInfo:
			l.levelAt = zap.InfoLevel
		default:
			l.levelAt = zap.DebugLevel
		}
	}
}

func (*Logger) WithHook(hook func(*zap.Logger, ...any) *zap.Logger, args ...any) func(*Logger) {
	return func(l *Logger) {
		l.core = hook(l.core, args...)
	}
}

func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.core.Debug(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	l.core.Error(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.core.Info(msg, fields...)
}
