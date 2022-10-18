package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/connector"
	"github.com/Mericusta/go-sgs/protocol"
)

const (
	DefaultServerAddress string = "127.0.0.1:6666"
	MaxConnectionCount   int    = 4096
	ChannelBuffer        int    = 8
)

type Server struct {
	listener  net.Listener
	linkerMgr []*Linker
}

// resource model 1: 1 - 1 - 3
// 1 client -> 1 socket -> 3 goroutine: read/write/logic
// resource model 2: 1 - 1 - 2 - 1/n
// 1 client -> 1 socket -> 2 goroutine: read/write -> logic: 1/n goroutine
// resource model 3: 1 - 1 - 1/n - 1/m
// 1 client -> 1 socket -> 1 goroutine: read -> logic: 1/n goroutine -> write: 1/m goroutine

// msg process
// os: tcp socket -> read goroutine: unpack []byte -> logic goroutine: unmarshal, handle
// logic goroutine: handle, marshal -> send goroutine: pack []byte -> os: tcp socket

// linker model
// SELECT 1: 相同 linker，在经过不同编译条件的情况下，可以处理不同格式的 packet
// NO-NEED 2: 相同 linker，在不经过编译的情况下，可以处理不同格式的 packet
// - 在没有“额外信息（如何处理 packet）”的情况下不知道 packet 的处理格式
// 	- 提前发包告知
// 	- 每个包内告知

// connector
// - unpack/pack tcp socket packet to []byte
// - unmarshal/marshal []byte to Msg

// dispatcher
// - receive Msg from recv goroutine
// - dispatch Msg to Handler and make Context
// - dispatch Msg to send goroutine by Linker
// - maybe different goroutine/program

// handler
// - handle Msg with Context
// - make Msg and Save Context

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

type Msg struct {
	id  protocol.ProtocolID
	msg protocol.Protocol
}

func (m *Msg) ID() protocol.ProtocolID {
	return m.id
}

func (m *Msg) Msg() protocol.Protocol {
	return m.msg
}

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
	}
}

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

func handleRecv(linker *Linker) {
	for {
		msgID, msg, err := linker.connector.RecvMsg()
		if err != nil {
			if err != io.EOF && err.(*net.OpError).Err != net.ErrClosed {
				// TODO: connector read error
				fmt.Printf("Error: connector read tcp socket packet occurs error: %v", err.Error())
				continue
			}
			// tcp socket closed
			close(linker.recv)
			linker.recv = nil
			return
		}
		linker.recv <- &Msg{
			id:  msgID,
			msg: msg,
		}
	}
}

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

func handleLogic(linker *Linker) {
	for {
		select {
		case msg, ok := <-linker.recv:
			if msg == nil || !ok {
				fmt.Printf("Error: linker logic goroutine receive msg is nil or not ok\n")
				continue
			}
			switch msg.ID() {
			case 1:
				fmt.Printf("Note: linker %v logic goroutine receive msg %v and response", linker.uid, msg.ID())
				// TODO: logic type assert and self-increase
				linker.send <- msg
			default:
				fmt.Printf("Error: linker %v logic goroutine receive unknown msg ID %v %v", linker.uid, msg.ID(), msg.Msg())
				continue
			}
		case <-linker.ctx.Done():
			fmt.Printf("Note: linker %v logic goroutine receive context done\n", linker.uid)
			close(linker.send)
			goto DONE
		}
	}
DONE:
	fmt.Printf("Note: linker %v logic goroutine done\n", linker.uid)
}
