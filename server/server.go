package server

import (
	"fmt"
	"net"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/protocol"
)

// protocol: a kind of data marshal/unmarshal algorithm
// a protocol always need a unique ID to make an instance
// - example:
// 	- json
// 	- protobuf
// "github.com/Mericusta/go-sgs/protocol"

// connector
// - unpack/pack tcp socket packet to []byte
// - unmarshal/marshal []byte to Msg
// 	- Msg has to implement unmarshal/marshal
// 	- Msg support different kinds unmarshal/marshal algorithm: proto, json, bson, messagepack...
// "github.com/Mericusta/go-sgs/connector"

// resource model 1: 1 - 1 - 3
// 1 client -> 1 socket -> 3 goroutine: read/write/logic
// resource model 2: 1 - 1 - 2 - 1/n
// 1 client -> 1 socket -> 2 goroutine: read/write -> logic: 1/n goroutine
// resource model 3: 1 - 1 - 1/n - 1/m
// 1 client -> 1 socket -> 1 goroutine: read -> logic: 1/n goroutine -> write: 1/m goroutine

// level 0: os tcp socket
// level 1: specific server program
// level 2: recv/send goroutine
// level 3: logic goroutine

// client - server transport process
// os: tcp socket -> read goroutine: unpack []byte -> logic goroutine: unmarshal, handle
// logic goroutine: handle, marshal -> send goroutine: pack []byte -> os: tcp socket
// ┌──────────────┬────────────────────────────────────┬─────────────────────┬─────────────────────────────┬──────────────────────────┐
// │      OS      │     recv goroutine: connector      │   recv goroutine    │ logic goroutine: dispatcher │ logic goroutine: handler │
// ├──────────────┼───────────────┬────────────────────┼─────────────────────┼─────────────────────────────┼──────────────────────────┤
// │  TCP Socket  │ unpack []byte │ unmarshal protocol │ recv channel <- Msg │     Msg <- recv channel     │        handle Msg        │
// └──────────────┴───────────────┴────────────────────┴─────────────────────┴─────────────────────────────┴──────────────────────────┘
// ┌──────────────────────────┬─────────────────────────────┬─────────────────────┬────────────────────────────────┬────────────┐
// │ logic goroutine: handler │ logic goroutine: dispatcher │   send goroutine    │   send goroutine: connector    │     OS     │
// ├──────────────────────────┼─────────────────────────────┼─────────────────────┼──────────────────┬─────────────┼────────────┤
// │         make Msg         │     send channel <- Msg     │ Msg <- send channel │ marshal protocol │ pack []byte │ TCP Socket │
// └──────────────────────────┴─────────────────────────────┴─────────────────────┴──────────────────┴─────────────┴────────────┘

// dispatcher
// - receive Msg from recv goroutine
// - dispatch Msg to Handler and make Context
// - dispatch Msg to send goroutine by Linker
// - maybe different goroutine/program

// handler
// - handler always need a unique ID (generally msg ID) to register callback
// - handle Msg with Context
// - make Msg and Save Context

// Server
type Server struct {
	listener  net.Listener
	linkerMgr []*linker.Linker
}

func New() *Server {
	listener, listenError := net.Listen("tcp", config.DefaultServerAddress)
	if listener == nil || listenError != nil {
		fmt.Printf("Error: listen tcp %v occurs error: %v\n", config.DefaultServerAddress, listenError.Error())
		return nil
	}

	return &Server{
		listener:  listener,
		linkerMgr: make([]*linker.Linker, 0, config.MaxConnectionCount),
	}
}

func (s *Server) Run() {
	for {
		connection, acceptError := s.listener.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				fmt.Printf("Error: server listener closed\n")
				return
			}
			fmt.Printf("Error: server listener accept connection occurs error: %v\n", acceptError.Error())
			continue
		}

		linker := linker.New(connection)
		go linker.HandleRecv()
		go linker.HandleSend()
		go handleLogic(linker)
		s.linkerMgr = append(s.linkerMgr, linker)
	}
}

func (s *Server) Exit(exitOvertimeSeconds int) {
	s.listener.Close()
}

// logic goroutine
func handleLogic(linker *linker.Linker) {
	msgCallbackMap := make(map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol))
	msgCallbackMap[MsgIDHeartBeatCounter] = func(linker *linker.Linker, msg protocol.Protocol) {
		heartBeatCounterMsg, ok := msg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", MsgIDHeartBeatCounter, msg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.send <- &msg.Msg{
			id:   MsgIDHeartBeatCounter,
			data: heartBeatCounterMsg,
		}
	}

	for {
		select {
		case msg, ok := <-linker.recv:
			if msg == nil || !ok {
				fmt.Printf("Error: linker logic goroutine receive msg is nil or not ok\n")
				continue
			}

			callback := msgCallbackMap[msg.ID()]
			if callback == nil {
				fmt.Printf("Error: msg ID %v callback is nil\n", msg.ID())
				continue
			}

			// TODO: make context
			callback(linker, msg.data)
		case <-linker.ctx.Done():
			fmt.Printf("Note: linker %v logic goroutine receive context done\n", linker.uid)
			close(linker.send)
			goto DONE
		}
	}
DONE:
	fmt.Printf("Note: linker %v logic goroutine done\n", linker.uid)
}
