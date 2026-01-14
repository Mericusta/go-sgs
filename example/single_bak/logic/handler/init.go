package handler

import (
	"net/http"

	"github.com/Mericusta/go-sgs/example/single/logic/router"
)

func Register() {
	router.RegisterWebsocketServerRoute(http.MethodGet, "/ws", dialHandler)
}
