package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/msg"
	"github.com/Mericusta/go-sgs/server"
)

func main() {
	counter := 10

	linkerMap := sync.Map{}
	wg := sync.WaitGroup{}
	wg.Add(counter)
	server := server.NewServer()
	go server.Run()

	for index := 0; index != counter; index++ {
		go func(i int) {
			connection, dialError := net.DialTimeout("tcp", config.DefaultServerAddress, time.Second)
			if dialError != nil {
				fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", i, config.DefaultServerAddress, dialError.Error())
				return
			}
			linker := linker.NewLinker(connection)
			linkerMap.Store(i, linker)
			go linker.HandleRecv()
			go linker.HandleSend()
			go func(l *linker.Linker, t int) {
				l.send <- msg.NewMsg(MsgIDHeartBeatCounter, &HeartBeatCounter{Count: t})
				s2cMsg, ok := <-l.recv
				if s2cMsg == nil || !ok {
					panic(fmt.Sprintf("%v %v", s2cMsg, ok))
				}
				if s2cMsg.ID() != MsgIDHeartBeatCounter {
					panic(s2cMsg.ID())
				}
				msg, ok := s2cMsg.data.(*HeartBeatCounter)
				if msg == nil || !ok {
					panic(fmt.Sprintf("%v %v", msg, ok))
				}
				if msg.Count != t+1 {
					panic(fmt.Sprintf("%v", msg.Count))
				}
				fmt.Printf("Note: client %v done %v\n", i, t)
				wg.Done()
			}(linker, i)
		}(index)
	}
	wg.Wait()
}
