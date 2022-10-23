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
	"github.com/Mericusta/go-sgs/link"
)

var (
	clientMgr          sync.Map
	clientCount        int
	clientContextMap   map[int]context.Context
	clientCancelMap    map[int]context.CancelFunc
	clientDialOvertime time.Duration
)

func init() {
	clientMgr = sync.Map{}
	clientCount = 1
	clientContextMap = make(map[int]context.Context)
	clientCancelMap = make(map[int]context.CancelFunc)
	clientDialOvertime = time.Second
}

func createClients(c int) {
	for index := 0; index != c; index++ {
		clientContextMap[index], clientCancelMap[index] = context.WithCancel(context.Background())
		go makeLink(index, config.DefaultServerAddress, clientDialOvertime)
	}
}

func makeLink(index int, addr string, overtime time.Duration) {
	connection, dialError := net.DialTimeout("tcp", addr, overtime)
	if dialError != nil {
		fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", index, config.DefaultServerAddress, dialError.Error())
		return
	}
	client := common.NewClient(link.New(connection), index)
	clientMgr.Store(index, client)
	fmt.Printf("Note: client %v created\n", index)
}

func runClients() {
	clientMgr.Range(func(key, value any) bool {
		clientIndex := key.(int)
		client := value.(*common.Client)
		go client.HandleRecv()
		go client.HandleSend()
		// go handleLogic(clientContextMap[clientIndex], client)
		go common.HandleLogic(clientContextMap[clientIndex], client, func(client *common.Client) {
			v1 := rand.Intn(int(time.Now().Unix()%86400)) + 1
			v2 := rand.Intn(int(time.Now().Unix()%86400)) + 1
			k := client.AddExpect(v1 + v2)
			client.Send(event.New(msg.C2SMsgID_CalculatorAdd, msg.C2SCalculatorData{Key: k, Value1: v1, Value2: v2}))
		}, clientMsgCallbackMap)
		return true
	})
}

// // data-driven
// func handleLogic(ctx context.Context, client *common.Client) {
// 	tickerCount := 0
// 	ticker := time.NewTicker(time.Millisecond * time.Duration(rand.Intn(100)+1))
// 	for {
// 		select {
// 		case <-ticker.C:
// 			v1 := rand.Intn(int(time.Now().Unix()%86400)) + 1
// 			v2 := rand.Intn(int(time.Now().Unix()%86400)) + 1
// 			client.data.expectMap[tickerCount] = v1 + v2
// 			tickerCount++
// 			client.Send(event.New(msg.C2SMsgID_CalculatorAdd, msg.C2SCalculatorData{Value1: v1, Value2: v2}))
// 		case e, ok := <-client.Recv():
// 			if e == nil || !ok {
// 				fmt.Printf("Error: client %v logic goroutine receive event is nil %v or not ok %v\n", client.UID(), e == nil, ok)
// 				continue
// 			}
// 			callback := clientMsgCallbackMap[e.ID()]
// 			if callback == nil {
// 				fmt.Printf("Error: event ID %v callback is nil\n", e.ID())
// 				continue
// 			}
// 			callback(client, e.Data())
// 		case <-ctx.Done():
// 			fmt.Printf("Note: client %v receive context done and end logic goroutine\n", client.UID())
// 			client.Exit()
// 			goto DONE
// 		}
// 	}
// DONE:
// 	fmt.Printf("Note: client %v logic goroutine done\n", client.UID())
// }

func main() {
	// register client protocol ID handler
	registerClientMsgCallback()

	// create clients
	createClients(clientCount)

	// run clients
	runClients()
}
