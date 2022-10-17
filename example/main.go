package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/connector"
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

type Linker struct {
	uid       uint64
	connector connector.Connector
	// msgMaker *MSG_MAKER_TYPE
	// recv     chan *msgPacket
	// send     chan *msgPacket
}

func NewLinker(c net.Conn) *Linker {
	return &Linker{
		Conn: connector.NewConnector(),
		uid:  uint64(time.Now().UnixNano()), // TODO: distributed-guid
		// msgMaker: &MSG_MAKER_TYPE{},
		// recv:     make(chan *msgPacket, ChannelBuffer),
		// send:     make(chan *msgPacket, ChannelBuffer),
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
		go handleRead(linker)
		go handleWrite(linker)
		go handleLogic(linker)
		s.linkerMgr = append(s.linkerMgr, linker)
	}
}

func (s *Server) Exit(exitOvertimeSeconds int) {
	s.listener.Close()
}

func handleRead(linker *Linker) {
	for {
		tag, data, err := unpackMsg(linker.Conn)
		if err != nil {
			if err != io.EOF && err.(*net.OpError).Err != net.ErrClosed {
				// TODO: os read error
				fmt.Printf("Error: os read tcp socket packet occurs error: %v", err.Error())
				continue
			}
			// tcp socket closed
			close(linker.recv)
			linker.recv = nil
			return
		}
		linker.recv <- &msgPacket{
			tag:  tag,
			data: data,
		}
	}
}

func handleWrite(connection net.Conn) {
	for {
		sendMsg, ok := <-r.SendChan
		if !ok {

			return
		}
		if sendMsg.MsgID != protocol.MSG_ID_C2S_HEART_BEAT {

		}
		err := r.Connector.SendMsg(sendMsg.MsgID, sendMsg.ByteData)
		if err != nil {
			if err.(*net.OpError).Err != net.ErrClosed {

			}
			return
		}
	}
}

func handleLogic(linker *Linker) {

}
