package moduleNet

import "github.com/gin-gonic/gin"

type HttpHandler func(IHttpContext)

type HttpRouteInfo struct {
	*gin.RouteInfo
	Handler HttpHandler
}
