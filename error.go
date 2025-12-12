package sgs

import (
	"go.uber.org/zap"
)

var (
	sendEventToSubjectError         string = "send event to subject occurs error"
	ErrorMsgHandleEventNonImplement string = "need to implement"
)

func SendEventToSubjectError(l *Logger, e IModuleEvent, err error) {
	l.Error(sendEventToSubjectError, zap.Any("toSubject", e.To()), zap.Any("data", e.Data()), zap.Error(err))
}
