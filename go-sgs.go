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
	root = root.New(append(mos, Base.WithCtx(rootCtx))...)
	sgsCanceler = rootCanceler
}

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
