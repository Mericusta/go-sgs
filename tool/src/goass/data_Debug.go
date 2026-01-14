package goass

import (
	"fmt"
	"time"

	"github.com/Mericusta/go-stp"
)

type data_Debug struct {
	logFilePath       string
	triggerCmdTickMap map[string]int
}

func newDebugData(logFilePath string) *data_Debug {
	err := stp.WriteFileByOverwriting(logFilePath, func(b []byte) ([]byte, error) {
		return nil, nil
	})
	if err != nil {
		panic(err)
	}
	return &data_Debug{logFilePath: logFilePath, triggerCmdTickMap: make(map[string]int)}
}

func (d *data_Debug) WriteLog(format string, args ...any) {
	if d == nil {
		return
	}
	err := stp.WriteFileByAppend(d.logFilePath, func(b []byte) ([]byte, error) {
		return []byte("[" + time.Now().Format(time.DateTime) + "] " + fmt.Sprintf(format, args...) + "\n"), nil
	})
	if err != nil {
		panic(err)
	}
}
