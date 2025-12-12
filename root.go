package sgs

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mericusta/go-stp"
	"go.uber.org/zap"
)

var root *ModuleRoot

// ModuleRoot 根模块
// - 一个进程中有且只能有一个 ModuleRoot
// - ModuleRoot 用来管理所有 channel
// - 合理利用 “goroutine 之间通过 channel 通信，而不需要关注通信的对象” 的特性解耦所有 module
// - 所有 goroutine 要有严格的父子关系
type ModuleRoot struct {
	// 组合基础模块
	ModuleBase
}

func (*ModuleRoot) New(mos ...ModuleOption) *ModuleRoot {
	m := &ModuleRoot{}
	for _, mo := range mos {
		mo(m)
	}
	return m
}

func (m *ModuleRoot) Hold() {
	// stp.Hold()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)
	for {
		s := <-c
		m.Logger().Debug("Hold, receive signal", zap.Any("signal", s))
		break
	}
	m.Logger().Debug("Hold, receive signal and exit")
}

func (m *ModuleRoot) Exit(identifies ...string) {
	m.Logger().Debug("Exit", zap.Any("manual unmount", identifies))
	// 优先卸载手动指定顺序的 module
	for _, identify := range identifies {
		err := m.Unmount(identify)
		if err != nil {
			m.Logger().Error("Exit, Unmount occurs error", zap.Error(err), zap.Any("subModuleIdentify", identify))
			continue
		}
	}
	// 卸载所有子 module
	identifiesArray := stp.NewArray(identifies)
	for _, subModuleIdentify := range m.Base().AllSubmodules() {
		if identifiesArray.Includes(subModuleIdentify) {
			continue
		}
		err := m.Unmount(subModuleIdentify)
		if err != nil {
			m.Logger().Error("Exit, Unmount occurs error", zap.Error(err), zap.Any("subModuleIdentify", subModuleIdentify))
			continue
		}
	}

	// TODO: 需要同步等待
	time.Sleep(time.Second * 3)
}

func Root() *ModuleRoot {
	return root
}

// 为单元测试生成的根节点
func (m *ModuleRoot) UnitTest(identify string) *ModuleRoot {
	Init(Base.WithIdentify(identify))
	return root
}
