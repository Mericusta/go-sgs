package moduleMonitor

// 使用接口的方式防止具体逻辑中直接引用该包

type IHandleReportBehavior interface {
	ReporterIdentify() string
	MessageType() string
	GateReceiveRequestMS() string
	GameReceiveRequestMS() string
	GameHandleRequestMS() string
	GateReceiveResponseMS() string
}

type ISubscribeRequestReportBehavior interface {
	SubscribeIdentify() string
}

type IRequestReportBehavior interface {
	SubscribeIdentify() string
	ReportIdentify() string
	MessageType() string
	DeltaMS() int64
	ErrorMsg() string
}
