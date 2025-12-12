package sgs

import (
	"sync"
	"sync/atomic"
)

type subjectManager struct {
	subjectChanMap      sync.Map
	subjectEventCounter atomic.Int64
}

var SubjectManager *subjectManager

func init() {
	SubjectManager = &subjectManager{
		subjectChanMap: sync.Map{},
	}
}

func (sm *subjectManager) LoadSubjectChan(identify string) (chan *ModuleEvent, bool) {
	ic, has := sm.subjectChanMap.Load(identify)
	if has {
		return ic.(chan *ModuleEvent), true
	}
	return nil, false
}

func (sm *subjectManager) LoadAndDeleteSubjectChan(identify string) (chan *ModuleEvent, bool) {
	ic, has := sm.subjectChanMap.LoadAndDelete(identify)
	if has {
		return ic.(chan *ModuleEvent), true
	}
	return nil, false
}

func (sm *subjectManager) LoadOrCreateSubjectChan(identify string, subjectMax int) chan *ModuleEvent {
	c := make(chan *ModuleEvent, subjectMax)
	ic, has := sm.subjectChanMap.LoadOrStore(identify, c)
	if has {
		return ic.(chan *ModuleEvent)
	}
	return c
}

func (sm *subjectManager) IncreaseEventCounter(delta int64) int64 {
	return sm.subjectEventCounter.Add(delta)
}

func (sm *subjectManager) DecreaseEventCounter(delta int64) int64 {
	return sm.subjectEventCounter.Add(-delta)
}

func (sm *subjectManager) LoadEventCounter() int64 {
	return sm.subjectEventCounter.Load()
}
