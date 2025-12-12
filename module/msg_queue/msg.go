package moduleMsgQueue

// 消息打包
type IMessagePatch interface {
	Pop() (any, bool)
}
