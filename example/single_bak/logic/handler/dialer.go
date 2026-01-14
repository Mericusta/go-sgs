package handler

import (
	"github.com/Mericusta/go-sgs"
	obj "github.com/Mericusta/go-sgs/example/single/logic/object"
	moduleNet "github.com/Mericusta/go-sgs/module/net"
	"github.com/gorilla/websocket"
)

func dialHandler(ctx moduleNet.IHttpContext) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(ctx.Raw().Writer, ctx.Raw().Request, nil)
	if err != nil {
		return
	}

	connection := &obj.Connection{
		RawConn: conn,
	}

	err = ctx.Module().Base().SendEvent(sgs.NewModuleEvent("gateServer", connection))
	if err != nil {
		return
	}
}
