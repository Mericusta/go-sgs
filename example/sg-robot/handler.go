package main

import (
	"math/rand"
	"time"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type RobotMgrHandler func(IRobotMgrContext, protocol.ProtocolMsg)

var robotMgrHandlerMrg map[protocol.ProtocolID]RobotMgrHandler

func RegisterRobotMgrHandler() {
	robotMgrHandlerMrg = make(map[protocol.ProtocolID]RobotMgrHandler)
	robotMgrHandlerMrg[msg.S2CMsgID_Login] = func(ctx IRobotMgrContext, p protocol.ProtocolMsg) {
		s2cMsg, ok := p.(*msg.S2CLoginData)
		if s2cMsg == nil || !ok {
			logger.Log().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Login), zap.Any("data", p))
			return
		}

		robot := model.NewRobot(ctx.UID())
		ctx.RobotMgr().Store(ctx.UID(), robot)

		logger.Log().Info("robot login", zap.Uint64("ID", robot.ID()))

		key := int(time.Now().UnixNano())
		v1, v2 := rand.Intn(1024), rand.Intn(1024)
		robot.AddExpect(key, v1+v2)
		c2sMsg := &msg.C2SBusinessData{
			Key: key, Value1: v1, Value2: v2,
		}
		ctx.Send(event.New(ctx.UID(), msg.C2SMsgID_Business, c2sMsg))
		logger.Log().Info("robot send business key value1 value2 wait expect", zap.Uint64("ID", robot.ID()), zap.Int("key", key), zap.Int("value1", v1), zap.Int("value2", v2), zap.Int("expect", v1+v2))
	}
}

type RobotHandler func(IRobotContext, protocol.ProtocolMsg)

var robotHandlerMgr map[protocol.ProtocolID]RobotHandler

func RegisterRobotHandler() {
	robotHandlerMgr = make(map[protocol.ProtocolID]RobotHandler)
	robotHandlerMgr[msg.S2CMsgID_Business] = func(ctx IRobotContext, p protocol.ProtocolMsg) {
		s2cMsg, ok := p.(*msg.S2CBusinessData)
		if s2cMsg == nil || !ok {
			logger.Log().Error("msg ID data not match", zap.Int("ID", msg.S2CMsgID_Business), zap.Any("data", p))
			return
		}

		if v, has := ctx.Robot().GetExpect(s2cMsg.Key); !has || v != s2cMsg.Result {
			logger.Log().Error("robot S2CMsgID_Business key result not match expect", zap.Uint64("ID", ctx.Robot().ID()), zap.Int("key", s2cMsg.Key), zap.Int("result", s2cMsg.Result), zap.Int("expect", v))
			return
		}

		ctx.Robot().DelExpect(s2cMsg.Key)
		logger.Log().Info("robot recv business key result, then delete expect", zap.Uint64("ID", ctx.Robot().ID()), zap.Int("key", s2cMsg.Key), zap.Int("result", s2cMsg.Result))

		time.Sleep(time.Second)

		if ctx.Robot().Counter() < requestCount {
			ctx.Robot().CounterIncrease()

			// condition: client/server exit passively
			key := int(time.Now().UnixNano())
			v1, v2 := rand.Intn(1024), rand.Intn(1024)
			ctx.Robot().AddExpect(key, v1+v2)
			c2sMsg := &msg.C2SBusinessData{
				Key: key, Value1: v1, Value2: v2,
			}
			ctx.Send(event.New(ctx.UID(), msg.C2SMsgID_Business, c2sMsg))
			logger.Log().Info("robot send business key value1 value2 wait expect", zap.Uint64("ID", ctx.Robot().ID()), zap.Int("key", key), zap.Int("value1", v1), zap.Int("value2", v2), zap.Int("expect", v1+v2))

			if ctx.Robot().Counter() == 6 {
				panic("robot painc here")
			}
		} else {
			// condition: client exit actively
			c2sMsg := &msg.C2SLogout{}
			ctx.Send(event.New(ctx.UID(), msg.C2SMsgID_Logout, c2sMsg))
			logger.Log().Info("robot send logout", zap.Uint64("ID", ctx.Robot().ID()))
		}
	}
}
