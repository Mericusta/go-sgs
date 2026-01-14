package sgs

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Mericusta/go-stp"
	"go.uber.org/zap"
)

var Base *ModuleBase

func (m *ModuleBase) Init(opts ...ModuleOption) {
	for _, opt := range opts {
		opt(m)
	}
}

// ModuleBase 基础模块
// 实现 Module 接口
type ModuleBase struct {
	// 模块名称
	identify string

	// 组合自己的 Module
	moduleSelf Module

	// 模块上下文
	ctx context.Context

	// 模块日志
	logger *Logger

	// 模块可处理事件数最大值
	handleEventMax int

	// --------

	// 状态自旋锁
	status atomic.Int32

	// 自旋锁 mount -> work 状态切换的等待自旋次数，用于定位
	statusMount2WorkSpinLockCount int64

	// 子 Module
	subModules *stp.CMap[string, *moduleController]

	// 事件生成标识
	eventCounter atomic.Int64
}

// setIdentify 设置模块标识
func (m *ModuleBase) setIdentify(i string) { m.identify = i }
func (m *ModuleBase) Identify() string     { return m.identify }

// setSelf 设置模块自引用
func (m *ModuleBase) setSelf(ms Module) { m.moduleSelf = ms }

// setCtx 设置模块上下文
func (m *ModuleBase) setCtx(ctx context.Context) { m.ctx = ctx }
func (m *ModuleBase) Ctx() context.Context       { return m.ctx }

// setLogger 设置模块日志
func (m *ModuleBase) setLogger(logger *Logger) { m.logger = logger }
func (m *ModuleBase) Logger() *Logger          { return m.logger }

// setHandleEventMax 设置模块可处理事件数最大值
func (m *ModuleBase) setHandleEventMax(max int) { m.handleEventMax = max }
func (m *ModuleBase) HandleEvent(*ModuleEvent)  { m.Logger().Debug("no implement") }

// Mounted 挂载后的回调函数，实现 Module 接口
func (m *ModuleBase) Mounted() {}

// Run 运行函数，实现 Module 接口
func (m *ModuleBase) Run() {}

// Unmounted 卸载后的回调函数，实现 Module 接口
func (m *ModuleBase) Unmounted() {}

// ----------------------------------------------------------------

func (m *ModuleBase) increaseStatusMount2WorkSpinLockCount() {
	m.statusMount2WorkSpinLockCount++
	if m.statusMount2WorkSpinLockCount%86400 == 0 {
		// spin 会暂停10ms，理论一天86400*100次，每86400次输出
		m.Logger().Info("waitEvent, but status is module_status_mount", zap.Any("count", m.statusMount2WorkSpinLockCount))
	}
}

var (
	ErrorInvalidEvent           = fmt.Errorf("invalid event")
	ErrorInvalidCallback        = fmt.Errorf("invalid callback")
	ErrorInvalidSubjectChanMap  = fmt.Errorf("invalid subject chan map")
	ErrorSubjectNotExists       = fmt.Errorf("subject not exists")
	ErrorSendEventCanceled      = fmt.Errorf("send event canceled")
	ErrorSendEventBlocked       = fmt.Errorf("send event blocked")
	ErrorSubscribeEventCanceled = fmt.Errorf("subscribe event canceled")
)

const (
	module_status_mount = iota
	module_status_work
	module_status_unmount
)

const (
	defaultSubjectMax     int    = 65535
	markEventBlockSubject string = ""
)

func matchMark(identify string) bool {
	return len(markEventBlockSubject) > 0 && (identify == markEventBlockSubject || strings.Contains(identify, markEventBlockSubject))
}

func (m *ModuleBase) subscribe() {
	handleEventMax := m.handleEventMax
	if handleEventMax <= 0 {
		handleEventMax = defaultSubjectMax
	}

	c := SubjectManager.LoadOrCreateSubjectChan(m.identify, handleEventMax)

	go m.waitEvent(c)
}

func (m *ModuleBase) waitEvent(c chan *ModuleEvent) {
	defer func() {
		if recoverInfo := recover(); recoverInfo != nil {
			m.Logger().Error("panic recover", zap.Any("info", recoverInfo))
			go m.waitEvent(c)
		}
	}()

	for index := 0; index != math.MaxInt64; index++ {
		switch m.status.Load() { // 自旋锁
		case module_status_mount:
			// m.Logger().Info("waitEvent, but status is module_status_mount")
			m.increaseStatusMount2WorkSpinLockCount()
			time.Sleep(time.Millisecond * 10)
		case module_status_work:
			select {
			case event, ok := <-c:
				if !ok {
					// 此时 chan 已经被关闭
					return
				}
				if event != nil {
					if matchMark(m.identify) {
						counter := SubjectManager.DecreaseEventCounter(1)
						m.logger.Debug("waitEvent, module_status_work, counter", zap.Any("fromIdentify", event.fromIdentify), zap.Any("counter", counter), zap.Any("id", event.id), zap.Any("data", event.data))
					}
					m.moduleSelf.HandleEvent(event)
				}
			case <-m.ctx.Done():
				m.logger.Debug("waitEvent, module_status_work, receive context canceler", zap.Any("counter", SubjectManager.LoadEventCounter()))
				return
			}
		case module_status_unmount:
			// 关闭后依次处理完剩下的内容
			for {
				select {
				case event, ok := <-c:
					if !ok {
						// 此时 chan 已经被关闭
						return
					}
					if event != nil {
						if matchMark(m.identify) {
							counter := SubjectManager.DecreaseEventCounter(1)
							m.logger.Debug("waitEvent, module_status_unmount, counter", zap.Any("fromIdentify", event.fromIdentify), zap.Any("counter", counter), zap.Any("id", event.id), zap.Any("data", event.data))
						}
						m.moduleSelf.HandleEvent(event)
					}
				case <-m.ctx.Done():
					m.logger.Debug("waitEvent, module_status_unmount, receive context canceler", zap.Any("counter", SubjectManager.LoadEventCounter()))
					return
				}
			}
		}
		// 让出 CPU，不然这里会无限循环导致 CPU 满负载卡死
		// 不知道这里 for Load 其他协程 CAS 达到的自旋锁效果，会不会有什么问题？
		runtime.Gosched()
	}

	m.Logger().Error("waitEvent, overload and need check it")
	go m.waitEvent(c)
}

func (m *ModuleBase) SendEvent(e *ModuleEvent) error {
	if e == nil || len(e.toSubject) == 0 || e.data == nil {
		// m.logger.Error("SendEvent, errInvalidEvent", zap.Any("toSubject", e.toSubject), zap.Any("id", e.id))
		return ErrorInvalidEvent
	}
	e.fromIdentify = m.identify
	e.id = m.eventCounter.Add(1) // 这里的 id 是指发送者累计发送的事件的 id，对于观察接收者来说，没有意义
	// m.logger.Debug("SendEvent", zap.Any("toSubject", e.toSubject), zap.Any("id", e.id), zap.Any("data", e.data))

	c, has := SubjectManager.LoadSubjectChan(e.toSubject)
	if !has {
		// m.logger.Error("SendEvent, errSubjectNotExists", zap.Any("toSubject", e.toSubject), zap.Any("id", e.id))
		return ErrorSubjectNotExists
	}
	select {
	case <-m.ctx.Done():
		m.logger.Debug("SendEvent, errSendEventCanceled", zap.Any("toSubject", e.toSubject), zap.Any("id", e.id))
		return ErrorSendEventCanceled
	case c <- e:
		if matchMark(m.identify) {
			counter := SubjectManager.IncreaseEventCounter(1)
			if counter > 0 && counter%128 == 0 {
				m.logger.Info("SendEvent, counter", zap.Any("toSubject", e.toSubject), zap.Any("counter", counter), zap.Any("id", e.id))
			}
		}
		// 往一个已经 close 的 chan 中发送数据会导致 panic
		// 这里采用 stack 形式的操作，以确保不会产生这种情况
		// stack：
		// - 创建 chan，添加到 chanMap，可被 SendEvent 查找到
		// - 从 chanMap 中移除，关闭 chan
		// m.logger.Debug("SendEvent, done", zap.Any("toSubject", e.toSubject), zap.Any("id", e.id))
		return nil
	default:
	}
	// 在这里一般表示 c 已经满了
	m.logger.Info("SendEvent, errSendEventBlocked", zap.Any("toSubject", e.toSubject), zap.Any("counter", SubjectManager.LoadEventCounter()), zap.Any("id", e.id))
	return ErrorSendEventBlocked
}

func (m *ModuleBase) GetSubModule(identify string) (Module, bool) {
	if m.subModules == nil {
		return nil, false
	}
	mc, has := m.subModules.Get(identify)
	if mc == nil || !has {
		return nil, false
	}
	return mc.Module(), true
}

func (m *ModuleBase) Mount(module Module) error {
	identify := module.Base().Identify()
	moduleCtx, moduleCanceler := context.WithCancel(m.ctx)
	module.Base().setCtx(moduleCtx)
	module.Base().setLogger(m.logger.New(
		Log.WithFields(zap.String("identify", identify)),
	))
	if m.subModules == nil {
		m.subModules = stp.NewCMap[string, *moduleController]()
	}
	// 首先开启订阅，可能存在并发开启的情况
	module.Base().subscribe()
	_, exists := m.subModules.Save(identify, &moduleController{
		module: module, canceler: moduleCanceler,
	})
	if exists {
		// 出现这种情况时表示并发 Mount
		// 开启了多个订阅协程，后 Mount 的关闭即可
		moduleCanceler()
		return nil
	}

	// 并发 Mount 时只会有一个 Mount 成功并且走到下面的逻辑
	// 其他触发 Mount 的逻辑会将消息放入 subject 中等待处理
	// 这里不需要担心会阻塞，因为 chan 是带有缓冲的

	// 但在这里还是不应该 HandleEvent
	// 因为处理逻辑可能会依赖 Mounted/Run 的结果
	// 所以需要状态控制，利用 ModuleBase 的 status
	// 此时 waiEvent 会进入自选状态，可以接收但不可以处理
	// 考虑到这种情况，Mounted 和 Run 都使用同步的方式

	// TODO: 都使用同步的话 Mounted 和 Run 可以整合在一起
	module.Mounted()

	// 这里需要使用异步的方式，因为需要 cas 以激活 waitEvent 协程
	go module.Run()

	// 更新 module 状态表示可以 HandleEvent
	// TODO: 这里 会有 cas 失败的情况吗？
	ok := module.Base().cas(module_status_mount, module_status_work)
	if !ok {
		for index := 0; index != 10; index++ {
			ok := module.Base().cas(module_status_mount, module_status_work)
			if ok {
				break
			}
			time.Sleep(time.Millisecond * 10)
			m.logger.Info("Mount, cas module_status_mount -> module_status_work failed", zap.Any("tryTimes", index))
		}
		return fmt.Errorf("Mount, overload and need check it")
	}
	// m.Logger().Debug("Mount, switch status", zap.Any("module", module.Base().identify), zap.Any("old", "module_status_mount"), zap.Any("new", "module_status_work"))

	return nil
}

func (m *ModuleBase) Unmount(identify string) error {
	if m.subModules == nil {
		return nil
	}

	mc, exists := m.subModules.Get(identify)
	if !exists {
		// 出现这种情况时表示并发 Unmount
		// 后 Unmount 的不处理即可
		m.logger.Debug("Unmount, submodule not exists", zap.Any("identify", identify))
		return nil
	}

	// 更新 module 状态表示不可以接收 event，并且会关闭 chan
	// 然后处理完 chan 中余下的内容，此时子 module 不再接受输入
	ok := mc.module.Base().cas(module_status_work, module_status_unmount)
	if !ok {
		// 出现这种情况时表示并发 Unmount
		// 后 Unmount 的不处理即可
		// m.logger.Debug("Unmount, cas module_status_work -> module_status_unmount failed")
		return nil
	}
	// m.Logger().Debug("Unmount, switch status", zap.Any("module", identify), zap.Any("old", "module_status_work"), zap.Any("new", "module_status_unmount"))

	// 在 waitEvent 中等待状态改变并且会引发死锁的问题所以在这里 remove 并且 close

	// 删除 subject 防止有新的 event
	// m.logger.Debug("Unmount, remove chan from subjectChanMap then close it", zap.Any("identify", identify))
	c, has := SubjectManager.LoadAndDeleteSubjectChan(identify)
	if has {
		// 然后关闭 chan
		close(c)
	}

	// 子 module 卸载完毕
	mc.module.Unmounted()

	// 删除 module
	m.subModules.Remove(identify)

	// 执行 module 的 canceler
	if mc.canceler != nil {
		mc.canceler()
	}

	return nil
}

func (m *ModuleBase) Module() Module {
	return m.moduleSelf
}

func (m *ModuleBase) cas(old, new int32) bool {
	return m.status.CompareAndSwap(old, new)
}

func (m *ModuleBase) Base() *ModuleBase {
	return m
}

func (m *ModuleBase) AllSubmodules() []string {
	submoduleIdentifySlice := make([]string, 0, 8)
	if m.subModules != nil {
		m.subModules.Range(func(s string, mc *moduleController) bool {
			submoduleIdentifySlice = append(submoduleIdentifySlice, s)
			return true
		})
	}
	return submoduleIdentifySlice
}
