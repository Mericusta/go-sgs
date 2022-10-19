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
var msgCallbackMap = make(map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol))

func registerMsgCallback() {
	msgCallbackMap[MsgIDHeartBeatCounter] = func(linker *linker.Linker, c2sMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := c2sMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", MsgIDHeartBeatCounter, c2sMsg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.Send(msg.New(MsgIDHeartBeatCounter, heartBeatCounterMsg))
	}
}

func main() {
	counter := 10
	wg := sync.WaitGroup{}
	wg.Add(counter)

	// server
	registerMsgCallback()
	serverCtx, serverCanceler := context.WithCancel(context.Background())
	server := server.New(dispatcher.New(msgCallbackMap))
	go server.Run(serverCtx)

	// client
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
			_linker := linker.New(connection)
			clientMap.Store(i, _linker)
			go _linker.HandleRecv()
			go _linker.HandleSend()
			go func(ctx context.Context, l *linker.Linker, t int) {
				l.Send(msg.New(MsgIDHeartBeatCounter, &HeartBeatCounter{Count: t}))
				s2cMsg, ok := l.Recv()
				if s2cMsg == nil || !ok {
					panic(fmt.Sprintf("%v %v", s2cMsg, ok))
				}
				if s2cMsg.ID() != MsgIDHeartBeatCounter {
					panic(s2cMsg.ID())
				}
				msg, ok := s2cMsg.Data().(*HeartBeatCounter)
				if msg == nil || !ok {
					panic(fmt.Sprintf("%v %v", msg, ok))
				}
				if msg.Count != t+1 {
					panic(fmt.Sprintf("%v", msg.Count))
				}
				fmt.Printf("Note: client %v %v done %v\n", i, l.UID(), t)
				wg.Done()
			}(ctx, _linker, i+1)
		}(ctx, index)
	}
	wg.Wait()

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	close(s)
	server.Exit()    // end tcp listener, all linker connection recv/send goroutine
	serverCanceler() // end logic goroutine
	time.Sleep(time.Second * 5)
}
