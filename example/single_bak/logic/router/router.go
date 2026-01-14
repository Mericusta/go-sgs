package router

import (
	moduleNet "github.com/Mericusta/go-sgs/module/net"
	"github.com/gin-gonic/gin"
)

var (
	wsServerDialRouteInfoMap = make(map[string]map[string]*moduleNet.HttpRouteInfo)
)

func GetWebsocketServerRouters() []*moduleNet.HttpRouteInfo {
	wsServerRoutes := make([]*moduleNet.HttpRouteInfo, 0, 16)
	for _, methodMap := range wsServerDialRouteInfoMap {
		for _, routeInfo := range methodMap {
			wsServerRoutes = append(wsServerRoutes, routeInfo)
		}
	}
	return wsServerRoutes
}

func RegisterWebsocketServerRoute(method, path string, handler moduleNet.HttpHandler) {
	if handler == nil || len(method) == 0 || len(path) == 0 {
		return
	}

	_, hasMethod := wsServerDialRouteInfoMap[method]
	if !hasMethod {
		wsServerDialRouteInfoMap[method] = make(map[string]*moduleNet.HttpRouteInfo)
	}
	_, hasPath := wsServerDialRouteInfoMap[method][path]
	if hasPath {
		return
	}
	wsServerDialRouteInfoMap[method][path] = &moduleNet.HttpRouteInfo{
		RouteInfo: &gin.RouteInfo{
			Method: method,
			Path:   path,
		},
		Handler: handler,
	}
}
