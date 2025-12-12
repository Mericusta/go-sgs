package sgs

// // 处理中间件
// type HandlerMiddleware interface {
// 	Do(*FrameworkEvent) bool
// }

// // recover 中间件
// type RecoverMiddleware interface {
// 	Recover(*FrameworkEvent, interface{}) bool
// }

type RunMiddleware interface {
	Do() bool
}
