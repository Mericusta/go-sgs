package moduleNet

import (
	"github.com/Mericusta/go-sgs"
	"github.com/gin-gonic/gin"
)

type IHttpContext interface {
	sgs.IModuleEventContext

	Raw() *gin.Context
}

type HttpContext struct {
	*ModuleHttpServer
	*gin.Context
}

func (ctx *HttpContext) Raw() *gin.Context {
	return ctx.Context
}

type WebsocketContext struct {
	*ModuleWebsocketServer
	*gin.Context
}

func (ctx *WebsocketContext) Raw() *gin.Context {
	return ctx.Context
}
