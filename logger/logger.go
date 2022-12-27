// go:build dev

package logger

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	var err error
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.OutputPaths = []string{"stdout"}
	logger, err = loggerCfg.Build()
	if err != nil {
		panic(err)
	}
}

func Log() *zap.Logger {
	return logger
}
