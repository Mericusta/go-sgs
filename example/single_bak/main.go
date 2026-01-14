package main

import (
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs"
	obj "github.com/Mericusta/go-sgs/example/single/logic/object"
	"github.com/Mericusta/go-sgs/example/single/logic/router"
	moduleNet "github.com/Mericusta/go-sgs/module/net"
	moduleTimingWheel "github.com/Mericusta/go-sgs/module/timing_wheel"
	"go.uber.org/zap"
)

func main() {
	var (
		identify = "single_server" // 进程主服务逻辑标识
		logLevel = sgs.LevelDebug
	)

	// 初始化框架
	sgs.Init(
		sgs.WithIdentify(identify),
		sgs.WithLogger(sgs.Log.New(
			sgs.Log.WithFields(zap.String("identify", identify)),
			sgs.Log.WithLevelAt(logLevel),
			sgs.Log.WithOutputPaths(fmt.Sprintf("%v.log", identify)),
		)),
	)

	// 待挂载的模块列表
	var modules []sgs.Module

	// 构造 时间轮服务 模块实例
	modules = append(modules, moduleTimingWheel.Service.New(
		sgs.WithIdentify("SingleServerTimingWheelService"),
		moduleTimingWheel.Service.WithTickerDuration(time.Millisecond*50),
	))

	// 构造 Websocket 服务 模块实例
	modules = append(modules, moduleNet.WebsocketServer.New(
		sgs.WithIdentify("SingleServerWebsocketService"),
		moduleNet.WebsocketServer.WithPort("8080"),
		moduleNet.WebsocketServer.WithRouters(router.GetWebsocketServerRouters()),
	))

	// 构造 进程主服务 实例
	modules = append(modules, obj.SingleServer.New(
		sgs.WithIdentify("SingleServer"),
	))

	// 挂载模块
	sgs.Mount(modules...)

	// 模块挂载完成
	sgs.Mounted()
	// 运行框架
	sgs.Run()
	// 挂起框架，等待退出信号
	sgs.Hold()
	// 退出框架
	sgs.Exit()
}
