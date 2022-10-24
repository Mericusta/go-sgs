package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/common"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/link"
)

type ClientMgr struct {
	*framework.Framework
	clientMgr          sync.Map
	clientContextMap   map[int]context.Context
	clientCancelMap    map[int]context.CancelFunc
	clientDialOvertime time.Duration
}

func NewClientMgr(cdo time.Duration) *ClientMgr {
	return &ClientMgr{
		Framework:          framework.New(),
		clientMgr:          sync.Map{},
		clientContextMap:   make(map[int]context.Context),
		clientCancelMap:    make(map[int]context.CancelFunc),
		clientDialOvertime: cdo,
	}
}

func (cm *ClientMgr) CreateClients(count int) {
	for index := 0; index != count; index++ {
		cm.clientContextMap[index], cm.clientCancelMap[index] = context.WithCancel(context.Background())
		go cm.makeLink(index, config.DefaultServerAddress, cm.clientDialOvertime)
	}
}

func (cm *ClientMgr) makeLink(index int, addr string, overtime time.Duration) {
	connection, dialError := net.DialTimeout("tcp", addr, overtime)
	if dialError != nil {
		fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", index, config.DefaultServerAddress, dialError.Error())
		return
	}
	client := common.NewClient(link.New(connection), index)
	cm.clientMgr.Store(index, client)
	fmt.Printf("Note: client %v created\n", index)
}

func (cm *ClientMgr) RunClients() {
	cm.clientMgr.Range(func(key, value any) bool {
		clientIndex := key.(int)
		client := value.(*common.Client)
		go client.HandleRecv()
		go client.HandleSend()
		go common.HandleLogic(cm.clientContextMap[clientIndex], client, func(client *common.Client) {
			v1 := rand.Intn(int(time.Now().Unix()%86400)) + 1
			v2 := rand.Intn(int(time.Now().Unix()%86400)) + 1
			k := client.AddExpect(v1 + v2)
			client.Send(event.New(msg.C2SMsgID_CalculatorAdd, msg.C2SCalculatorData{Key: k, Value1: v1, Value2: v2}))
		}, clientMsgCallbackMap)
		return true
	})
}
