package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

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
	const robotCount int = 1

	// register protocol ID handler
	RegisterRobotHandler()

	// create robot manager
	robotMgr := NewRobotMgr()

	// create clients
	robotMgr.CreateRobots(robotCount)

	// // run clients
	// runClients()

	// watch system signal
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	robotMgr.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
