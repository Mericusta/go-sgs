package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/connector"
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

// Msg: contain protocol unique ID and protocol instance to transport from recv/send goroutine to logic goroutine
type Msg struct {
	id   protocol.ProtocolID
	data protocol.Protocol
}

func (m *Msg) ID() protocol.ProtocolID {
	return m.id
}

func (m *Msg) Msg() protocol.Protocol {
	return m.data
}

const (
	DefaultServerAddress string = "127.0.0.1:6666"
	MaxConnectionCount   int    = 4096
	ChannelBuffer        int    = 8
)

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

// Server
type Server struct {
	listener  net.Listener
	linkerMgr []*Linker
}

// Linker
// SELECT 1: 相同 linker，在经过不同编译条件的情况下，可以处理不同格式的 packet
// NO-NEED 2: 相同 linker，在不经过编译的情况下，可以处理不同格式的 packet
// - 在没有“额外信息（如何处理 packet）”的情况下不知道 packet 的处理格式
// 	- 提前发包告知
// 	- 每个包内告知

type Linker struct {
	uid       uint64
	connector connector.Connector
	recv      chan *Msg
	send      chan *Msg
	ctx       context.Context // dispatcher make
}

func NewLinker(connection net.Conn) *Linker {
	return &Linker{
		connector: connector.NewConnector(connection),
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		recv:      make(chan *Msg, ChannelBuffer),
		send:      make(chan *Msg, ChannelBuffer),
		ctx:       context.Background(),
	}
}

// dispatcher
// - receive Msg from recv goroutine
// - dispatch Msg to Handler and make Context
// - dispatch Msg to send goroutine by Linker
// - maybe different goroutine/program

// handler
// - handler always need a unique ID (generally msg ID) to register callback
// - handle Msg with Context
// - make Msg and Save Context

func NewServer() *Server {
	listener, listenError := net.Listen("tcp", DefaultServerAddress)
	if listener == nil || listenError != nil {
		fmt.Printf("Error: listen tcp %v occurs error: %v\n", DefaultServerAddress, listenError.Error())
		return nil
	}

	return &Server{
		listener:  listener,
		linkerMgr: make([]*Linker, 0, MaxConnectionCount),
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

		linker := NewLinker(connection)
		go handleRecv(linker)
		go handleSend(linker)
		go handleLogic(linker)
		s.linkerMgr = append(s.linkerMgr, linker)
	}
}

func (s *Server) Exit(exitOvertimeSeconds int) {
	s.listener.Close()
}

// recv goroutine
func handleRecv(linker *Linker) {
	for {
		msgID, msg, err := linker.connector.RecvMsg()
		if err != nil {
			fmt.Printf("Error: connector read tcp socket packet occurs error: %v\n", err.Error())
			if err != io.EOF {
				if opError, ok := err.(*net.OpError); ok && opError.Err != net.ErrClosed {
					// TODO: connector read error
					continue
				}
			}
			// tcp socket closed
			close(linker.recv)
			linker.recv = nil
			return
		}
		linker.recv <- &Msg{
			id:   msgID,
			data: msg,
		}
	}
}

// send goroutine
func handleSend(linker *Linker) {
	for {
		sendMsg, ok := <-linker.send
		if !ok {
			fmt.Printf("Error: send msg is not ok\n")
			continue
		}
		err := linker.connector.SendMsg(sendMsg.ID(), sendMsg.Msg())
		if err != nil {
			// TODO: connector send error
			fmt.Printf("Error: connector send tcp socket packet occurs error: %v", err.Error())
		}
	}
}

// logic goroutine
func handleLogic(linker *Linker) {
	msgCallbackMap := make(map[protocol.ProtocolID]func(*Linker, protocol.Protocol))
	msgCallbackMap[MsgIDHeartBeatCounter] = func(linker *Linker, msg protocol.Protocol) {
		heartBeatCounterMsg, ok := msg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", MsgIDHeartBeatCounter, msg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.send <- &Msg{
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

func main() {
	counter := 10

	linkerMap := sync.Map{}
	wg := sync.WaitGroup{}
	wg.Add(counter)
	server := NewServer()
	go server.Run()

	for index := 0; index != counter; index++ {
		go func(i int) {
			connection, dialError := net.DialTimeout("tcp", DefaultServerAddress, time.Second)
			if dialError != nil {
				fmt.Printf("Error: client %v dial tcp address %v occurs error: %v", i, DefaultServerAddress, dialError.Error())
				return
			}
			linker := NewLinker(connection)
			linkerMap.Store(i, linker)
			go handleRecv(linker)
			go handleSend(linker)
			go func(l *Linker, t int) {
				l.send <- &Msg{
					id:   MsgIDHeartBeatCounter,
					data: &HeartBeatCounter{Count: t},
				}
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
