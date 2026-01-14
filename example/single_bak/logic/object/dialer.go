package obj

import "github.com/gorilla/websocket"

type Connection struct {
	RawConn *websocket.Conn
}
