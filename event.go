package sgs

type EventIdentify int

const (
	EVENT_CONTROLLER_MOUNT_MODULE EventIdentify = iota + 1
	EVENT_CONTROLLER_UNMOUNT_MODULE
	EVENT_MODULE_CALL
	EVENT_TYPE_MODULE_RUN
	EVENT_TYPE_MODULE_CALL
	EVENT_TYPE_MODULE_EXIT
)

// var eventTypeKeywordMap map[EventIdentify]string = map[EventIdentify]string{
// 	EVENT_CONTROLLER_MOUNT_MODULE:   "controller_mount_module",
// 	EVENT_CONTROLLER_UNMOUNT_MODULE: "controller_unmount_module",
// 	EVENT_MODULE_CALL:               "module_call",
// }

type IModuleEvent interface {
	From() string
	To() string
	Data() any
}

type ModuleEvent struct {
	fromIdentify string
	toSubject    string
	data         any
	id           int64
}

func NewModuleEvent(subject string, data any) *ModuleEvent {
	return &ModuleEvent{
		toSubject: subject,
		data:      data,
	}
}

func (me *ModuleEvent) From() string { return me.fromIdentify }
func (me *ModuleEvent) To() string   { return me.toSubject }
func (me *ModuleEvent) Data() any    { return me.data }
