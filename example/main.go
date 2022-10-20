package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/msg"
	"github.com/Mericusta/go-sgs/protocol"
	"github.com/Mericusta/go-sgs/server"
)

// business logic
// TODO: callback 使用 ctx 传递上下文，而不是直接传递 linker？
// TODO: callback 直接传递 Client 包裹 linker？
var c2sMsgCallbackMap = make(map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol))

func registerC2SMsgCallback() {
	c2sMsgCallbackMap[C2SMsgID_HeartBeatCounter] = func(linker *linker.Linker, c2sMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := c2sMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", C2SMsgID_HeartBeatCounter, c2sMsg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.Send(msg.New(S2CMsgID_HeartBeatCounter, heartBeatCounterMsg))
	}

	s2cMsgCallbackMap[S2CMsgID_HeartBeatCounter] = func(linker *linker.Linker, s2cMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := s2cMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", C2SMsgID_HeartBeatCounter, s2cMsg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.Send(msg.New(C2SMsgID_HeartBeatCounter, heartBeatCounterMsg))
	}
}

func main() {
	counter := 1
	wg := sync.WaitGroup{}
	wg.Add(counter)

	// server
	registerC2SMsgCallback()
	serverCtx, serverCanceler := context.WithCancel(context.Background())
	server := server.New(dispatcher.New(c2sMsgCallbackMap))
	go server.Run(serverCtx)

	// client
	registerS2CMsgCallback()
	clientMap := sync.Map{}
	clientCancelMap := make(map[int]context.CancelFunc)
	for index := 0; index != counter; index++ {
		var ctx context.Context
		ctx, clientCancelMap[index] = context.WithCancel(context.Background())
		go func(ctx context.Context, i int) {
			connection, dialError := net.DialTimeout("tcp", config.DefaultServerAddress, time.Second)
			if dialError != nil {
				fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", i, config.DefaultServerAddress, dialError.Error())
				return
			}
			client := NewClient(linker.New(connection))
			clientMap.Store(i, client)
			fmt.Printf("Note: client %v create linker %v\n", i, client.UID())
			go client.HandleRecv()
			go client.HandleSend()

			// go func(ctx context.Context, l *linker.Linker, t int) {
			// 	for {
			// 		select {
			// 		case msg, ok := <-l.recv
			// 		}
			// 	}

			// 	l.Send(msg.New(MsgIDHeartBeatCounter, &HeartBeatCounter{Count: t}))

			// 	s2cMsg, ok := l.Recv()
			// 	if s2cMsg == nil || !ok {
			// 		panic(fmt.Sprintf("%v %v", s2cMsg, ok))
			// 	}
			// 	if s2cMsg.ID() != MsgIDHeartBeatCounter {
			// 		panic(s2cMsg.ID())
			// 	}
			// 	msg, ok := s2cMsg.Data().(*HeartBeatCounter)
			// 	if msg == nil || !ok {
			// 		panic(fmt.Sprintf("%v %v", msg, ok))
			// 	}
			// 	if msg.Count != t+1 {
			// 		panic(fmt.Sprintf("%v", msg.Count))
			// 	}
			// 	fmt.Printf("Note: client %v linker %v done %v\n", i, _linker.UID(), t)
			// 	wg.Done()
			// }(ctx, _linker, i+1)

			go client.HandleLogic(ctx, s2cMsgCallbackMap)
		}(ctx, index)
	}
	wg.Wait()

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	server.Exit() // end tcp listener, all linker connection recv goroutine
	fmt.Printf("Note: execute canceler\n")
	serverCanceler() // end logic goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
