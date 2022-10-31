package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/example/model/robot"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/link"
)

type RobotMgr struct {
	f *framework.Framework
	m *sync.Map
}

func NewRobotMgr() *RobotMgr {
	return &RobotMgr{
		f: framework.New(),
		m: &sync.Map{},
	}
}

func (rm *RobotMgr) CreateRobots(count int) {
	for index := 0; index != count; index++ {
		go rm.makeLink(robot.NewRobot(uint64(index)), config.DefaultServerAddress)
	}
}

func (rm *RobotMgr) makeLink(r *robot.Robot, addr string) {
	connection, dialError := net.DialTimeout("tcp", addr, config.TcpDialOvertime)
	if dialError != nil {
		fmt.Printf("Error: robot %v dial tcp address %v occurs error: %v", r.ID(), config.DefaultServerAddress, dialError.Error())
		return
	}
	l := link.New(connection)
	r.Dispatcher = dispatcher.New(l)
	rm.m.Store(r.ID(), r)
	fmt.Printf("Note: robot %v created\n", r.ID())
}

func (rm *RobotMgr) RunClients() {
	rm.m.Range(func(key, value any) bool {
		// robotID := key.(int)
		robot := value.(*robot.Robot)
		go robot.Link().HandleRecv()
		go robot.Link().HandleSend()
		// go common.HandleLogic(cm.clientContextMap[clientIndex], client, func(client *common.Client) {
		// 	v1 := rand.Intn(int(time.Now().Unix()%86400)) + 1
		// 	v2 := rand.Intn(int(time.Now().Unix()%86400)) + 1
		// 	k := client.AddExpect(v1 + v2)
		// 	client.Send(event.New(msg.C2SMsgID_CalculatorAdd, msg.C2SCalculatorData{Key: k, Value1: v1, Value2: v2}))
		// }, clientMsgCallbackMap)
		return true
	})
}

func (rm *RobotMgr) Exit() {

}
