package sgs

// ModuleOption 模块选项
type ModuleOption func(Module)

// ModuleConstructor 模块构造函数
type ModuleConstructor func(...ModuleOption) Module

// Module 基础模块
type Module interface {
	// 由 ModuleBase 实现并提供的接口，对外暴漏，但不需要应用层实现
	Base() *ModuleBase

	// 模块需要实现的接口，可能是并发调用的

	// 模块挂载后的回调函数，通常需要被实现
	Mounted()

	// 模块卸载后的回调函数，通常需要被实现
	Unmounted()

	// 模块处理接收到的事件，通常需要被实现
	HandleEvent(*ModuleEvent)

	// 模块的主要运行函数，通常需要被实现
	Run()
}
