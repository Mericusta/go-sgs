package sgs

import (
	"context"

	"go.uber.org/zap"
)

var sgsCanceler context.CancelFunc

// Init main 协程：初始化框架
func Init(mos ...ModuleOption) {
	// 构造根模块
	rootCtx, rootCanceler := context.WithCancel(context.Background())
	root = root.new(append(mos, WithCtx(rootCtx))...)
	sgsCanceler = rootCanceler
}

// WithIdentify 设置模块标识选项
func WithIdentify(i string) ModuleOption {
	return func(m Module) {
		m.Base().setIdentify(i)
		m.Base().setSelf(m)
	}
}

// WithCtx 设置模块上下文选项
func WithCtx(c context.Context) ModuleOption {
	return func(m Module) { m.Base().setCtx(c) }
}

// WithLogger 设置模块日志选项
func WithLogger(l *Logger) ModuleOption {
	return func(m Module) {
		m.Base().setLogger(l.New(
			Log.WithFields(zap.String("identify", m.Base().Identify())),
		))
	}
}

// WithHandleEventMax 设置模块可处理事件数最大值选项
func WithHandleEventMax(max int) ModuleOption {
	return func(m Module) { m.Base().setHandleEventMax(max) }
}

// Mount main 协程：框架挂载模块
func Mount(modules ...Module) {
	for _, module := range modules {
		root.Logger().Debug("OBSERVE: Mount module", zap.Any("module", module.Base().Identify()))
		err := root.Mount(module)
		if err != nil {
			// 根 module 挂载子 module 失败不提供服务
			panic(err)
		}
	}
}

// Mounted main 协程：框架模块挂载完成
func Mounted() {
	root.Logger().Debug("OBSERVE: Mounted")
	root.Mounted()
}

// Unmount main 协程：框架卸载模块
func Unmount(m ...Module) {

}

// Run main 协程：运行框架
func Run() {
	// 运行框架主逻辑
	go root.Run()
}

// Hold main 协程：挂起框架
func Hold() {
	root.Hold()
}

// Exit main 协程：退出框架
func Exit(identifies ...string) {
	root.Exit(identifies...)
	sgsCanceler()
}
