package main

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"os"
// 	"os/signal"
// 	"sync"
// 	"time"

// 	"github.com/Mericusta/go-sgs/config"
// 	"github.com/Mericusta/go-sgs/link"
// 	"github.com/Mericusta/go-sgs/server"
// )

// func main() {
// 	counter := 1
// 	wg := sync.WaitGroup{}
// 	wg.Add(counter)

// 	// client
// 	registerS2CMsgCallback()
// 	clientMap := sync.Map{}
// 	clientCancelMap := make(map[int]context.CancelFunc)
// 	for index := 0; index != counter; index++ {
// 		var ctx context.Context
// 		ctx, clientCancelMap[index] = context.WithCancel(context.Background())
// 		go func(ctx context.Context, i int) {
// 			connection, dialError := net.DialTimeout("tcp", config.DefaultServerAddress, time.Second)
// 			if dialError != nil {
// 				fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", i, config.DefaultServerAddress, dialError.Error())
// 				return
// 			}
// 			client := NewClient(link.New(connection))
// 			clientMap.Store(i, client)
// 			fmt.Printf("Note: client %v create link %v\n", i, client.UID())
// 			go client.HandleRecv()
// 			go client.HandleSend()

// 			// go func(ctx context.Context, l *link.Linker, t int) {
// 			// 	for {
// 			// 		select {
// 			// 		case msg, ok := <-l.recv
// 			// 		}
// 			// 	}

// 			// 	l.Send(msg.New(MsgIDHeartBeatCounter, &HeartBeatCounter{Count: t}))

// 			// 	s2cMsg, ok := l.Recv()
// 			// 	if s2cMsg == nil || !ok {
// 			// 		panic(fmt.Sprintf("%v %v", s2cMsg, ok))
// 			// 	}
// 			// 	if s2cMsg.ID() != MsgIDHeartBeatCounter {
// 			// 		panic(s2cMsg.ID())
// 			// 	}
// 			// 	msg, ok := s2cMsg.Data().(*HeartBeatCounter)
// 			// 	if msg == nil || !ok {
// 			// 		panic(fmt.Sprintf("%v %v", msg, ok))
// 			// 	}
// 			// 	if msg.Count != t+1 {
// 			// 		panic(fmt.Sprintf("%v", msg.Count))
// 			// 	}
// 			// 	fmt.Printf("Note: client %v link %v done %v\n", i, _linker.UID(), t)
// 			// 	wg.Done()
// 			// }(ctx, _linker, i+1)

// 			go client.HandleLogic(ctx, s2cMsgCallbackMap)
// 		}(ctx, index)
// 	}
// 	wg.Wait()

// 	s := make(chan os.Signal)
// 	signal.Notify(s, os.Interrupt)
// 	<-s
// 	fmt.Printf("Note: close signal\n")
// 	close(s)
// 	fmt.Printf("Note: server exit\n")
// 	server.Exit() // end tcp listener, all link connection recv goroutine
// 	fmt.Printf("Note: execute canceler\n")
// 	serverCanceler() // end logic goroutine
// 	fmt.Printf("Note: waitting 5 seconds\n")
// 	time.Sleep(time.Second * 5)
// }
