package moduleTimingWheel

import (
	"slices"
	"sync/atomic"
	"time"

	"github.com/Mericusta/go-sgs"
	"go.uber.org/zap"
)

type selfTick struct{}

type ModuleTimingWheel struct {
	sgs.ModuleBase

	ticker                         *time.Ticker
	receiver                       chan ITickHandler
	triggerTodoSlice               []ITickHandler
	lessThanMinuteTodoSlice        []ITickHandler
	moreThanMinuteTodoSlice        []ITickHandler
	blockedTodoBehaviorSlice       []ITickHandler // 用来存放 handleBehavior 中由于 receiver 阻塞住的行为
	blockedTodoSliceTimer          *time.Timer    // 动态定时器
	receiveBehaviorOvertimeSeconds int            // 动态最大延迟秒数，根据 blockedTodoSlice 的长度决定，默认5s，数量每增加10减少1s
	headTail                       atomic.Uint64  // TODO: 头尾指针，参考 sync.poolDequeue
}

var Service *ModuleTimingWheel

func (*ModuleTimingWheel) New(mos ...sgs.ModuleOption) *ModuleTimingWheel {
	mtw := &ModuleTimingWheel{}
	for _, mo := range mos {
		mo(mtw)
	}
	return mtw
}

func (*ModuleTimingWheel) WithTickerDuration(d time.Duration) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleTimingWheel).ticker = time.NewTicker(d) }
}

func (m *ModuleTimingWheel) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted")

	if m.ticker == nil {
		m.ticker = time.NewTicker(time.Second)
	}
	m.receiver = make(chan ITickHandler, 1024)
	m.triggerTodoSlice = make([]ITickHandler, 0, 1024)
	m.lessThanMinuteTodoSlice = make([]ITickHandler, 0, 1024)
	m.moreThanMinuteTodoSlice = make([]ITickHandler, 0, 1024)
	m.blockedTodoBehaviorSlice = make([]ITickHandler, 0, 1024)
	m.receiveBehaviorOvertimeSeconds = 5
	m.blockedTodoSliceTimer = time.NewTimer(time.Second * time.Duration(m.receiveBehaviorOvertimeSeconds))
}

func (m *ModuleTimingWheel) HandleEvent(event *sgs.ModuleEvent) {
	switch data := event.Data().(type) {
	case ITickHandler:
		m.handleBehavior(data)
	case *selfTick:
		m.handleBehavior(nil)
	default:
		m.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
	}
}

func (m *ModuleTimingWheel) handleBehavior(behavior ITickHandler) {
	if len(m.blockedTodoBehaviorSlice) > 0 { // 缓存列表已经存在了，则先丢到缓存列表中
		if behavior != nil {
			m.blockedTodoBehaviorSlice = append(m.blockedTodoBehaviorSlice, behavior)
		}
		// 超时时间可以容忍，即 receiver 未阻塞很严重导致缓存列表中行为没有超过阈值
		// 超时时间不可以容忍，则只添加到缓存列表中
		if m.receiveBehaviorOvertimeSeconds >= 0 {
			for index, blockedTodoBehavior := range m.blockedTodoBehaviorSlice { // 然后逐个将缓存列表中的行为丢到主逻辑中
				m.blockedTodoSliceTimer.Reset(time.Second * time.Duration(m.receiveBehaviorOvertimeSeconds))
				select {
				case m.receiver <- blockedTodoBehavior: // 成功则继续下一个
					continue // TODO: 这里是否需要动态更新 m.receiveBehaviorOvertimeSeconds
				case <-m.blockedTodoSliceTimer.C: // 超时
					// 修改已处理的数据，并且更新超时延迟
					m.blockedTodoBehaviorSlice = m.blockedTodoBehaviorSlice[index:]
					m.updateBlockedTodoSliceDurationSeconds()
					return // 从这里结束的是缓存队列未处理完
				}
			}
			// 从这里结束的是缓存队列已处理完
			m.blockedTodoBehaviorSlice = make([]ITickHandler, 0, 1024)
			m.updateBlockedTodoSliceDurationSeconds()
			// 日志点2：结束阻塞情况
			m.Logger().Info("handleBehavior, blocked todo behavior slice is clear")
		} else {
			// 报警
			// 日志点3：阻塞情况非常严重
			m.Logger().Info("handleBehavior, receive behavior overtime seconds <= 0", zap.Any("count", len(m.blockedTodoBehaviorSlice)))
		}
	} else { // 缓存列表为空，则优先丢到主逻辑中
		if behavior != nil {
			m.blockedTodoSliceTimer.Reset(time.Second * time.Duration(m.receiveBehaviorOvertimeSeconds))
			select {
			case m.receiver <- behavior: // 成功
			case <-m.blockedTodoSliceTimer.C: // 超时
				// 丢到缓存队列中，并且更新超时延迟，报警
				m.blockedTodoBehaviorSlice = append(m.blockedTodoBehaviorSlice, behavior)
				m.updateBlockedTodoSliceDurationSeconds()
				time.Sleep(time.Millisecond * 100)
				err := m.SendEvent(sgs.NewModuleEvent(m.Identify(), &selfTick{}))
				if err != nil {
					m.Logger().Error("handleBehavior, handle behavior overtime, self tick occurs error", zap.Error(err))
				}
				// 日志点1：出现阻塞情况
				m.Logger().Info("handleBehavior, handle behavior overtime, append to blocked todo behavior slice", zap.Any("count", len(m.blockedTodoBehaviorSlice)), zap.Any("overtime seconds", m.receiveBehaviorOvertimeSeconds))
			}
		}
	}
}

func (m *ModuleTimingWheel) Run() {
	for {
		select {
		case <-m.ticker.C:
			now := time.Now()
			// less
			lessThanMinuteDone := true
			for index, todoBehavior := range m.lessThanMinuteTodoSlice {
				if todoBehavior.At().Sub(now).Seconds() > 1 {
					lessThanMinuteDone = false
					m.lessThanMinuteTodoSlice = m.lessThanMinuteTodoSlice[index:]
					break
				}
				m.triggerTodoSlice = append(m.triggerTodoSlice, m.lessThanMinuteTodoSlice[index])
			}
			if lessThanMinuteDone {
				m.lessThanMinuteTodoSlice = nil
			}
			// trigger todo
			for _, behavior := range m.triggerTodoSlice {
				// m.Logger().Debug("Run, trigger", zap.Any("at", behavior.At().Unix()), zap.Any("len(triggerTodoSlice)", len(m.triggerTodoSlice)), zap.Any("len(m.lessThanMinuteTodoSlice)", len(m.lessThanMinuteTodoSlice)))
				behavior.Trigger(m)
			}
			m.triggerTodoSlice = nil
			// more
			moreThanMinuteDone := true
			for index, todoBehavior := range m.moreThanMinuteTodoSlice {
				if todoBehavior.At().Sub(now).Minutes() > 1 {
					moreThanMinuteDone = false
					m.moreThanMinuteTodoSlice = m.moreThanMinuteTodoSlice[index:]
					break
				}
				m.lessThanMinuteTodoSlice = append(m.lessThanMinuteTodoSlice, m.moreThanMinuteTodoSlice[index])
			}
			if moreThanMinuteDone {
				m.moreThanMinuteTodoSlice = nil
			}
		case newBehavior := <-m.receiver:
			behaviorFromNow := time.Until(newBehavior.At())
			if behaviorFromNow.Seconds() < 1 {
				// m.Logger().Debug("Run, less second", zap.Any("at", newBehavior.At().Unix()), zap.Any("len(triggerTodoSlice)", len(m.triggerTodoSlice)))
				m.triggerTodoSlice = append([]ITickHandler{newBehavior}, m.triggerTodoSlice...)
			} else if behaviorFromNow.Minutes() < 1 {
				insert := false
				for index, todoBehavior := range m.lessThanMinuteTodoSlice {
					if newBehavior.At().Sub(todoBehavior.At()).Seconds() <= 1 {
						insert = true
						m.lessThanMinuteTodoSlice = slices.Insert(m.lessThanMinuteTodoSlice, index, newBehavior)
						// m.Logger().Debug("Run, less than minute", zap.Any("at", newBehavior.At().Unix()), zap.Any("len(lessThanMinuteTodoSlice)", len(m.lessThanMinuteTodoSlice)), zap.Any("insert index", index))
						break
					}
					continue
				}
				if !insert {
					m.lessThanMinuteTodoSlice = append(m.lessThanMinuteTodoSlice, newBehavior)
					// m.Logger().Debug("Run, less than minute", zap.Any("at", newBehavior.At().Unix()), zap.Any("len(lessThanMinuteTodoSlice)", len(m.lessThanMinuteTodoSlice)), zap.Any("insert index", -1))
				}
			} else {
				insert := false
				for index, todoBehavior := range m.moreThanMinuteTodoSlice {
					if newBehavior.At().Sub(todoBehavior.At()).Seconds() <= 1 {
						insert = true
						m.moreThanMinuteTodoSlice = slices.Insert(m.moreThanMinuteTodoSlice, index, newBehavior)
						break
					}
					continue
				}
				if !insert {
					m.moreThanMinuteTodoSlice = append(m.moreThanMinuteTodoSlice, newBehavior)
				}
			}
		}
	}
}

func (m *ModuleTimingWheel) updateBlockedTodoSliceDurationSeconds() {
	releaseCount := len(m.blockedTodoBehaviorSlice)
	decreaseSeconds := releaseCount / 10
	if decreaseSeconds >= 5 {
		m.receiveBehaviorOvertimeSeconds = 0
	} else {
		m.receiveBehaviorOvertimeSeconds -= decreaseSeconds
	}
}
