package moduleNet

import (
	"errors"
	"net/http"

	"github.com/Mericusta/go-sgs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 监听端口，建立 socket 链接
type ModuleWebsocketServer struct {
	sgs.ModuleBase

	// 监听地址
	port string

	// dial 地址
	routers []*HttpRouteInfo

	// gin 引擎
	engine *gin.Engine

	// http server
	server *http.Server
}

var WebsocketServer *ModuleWebsocketServer

func (*ModuleWebsocketServer) New(mos ...sgs.ModuleOption) *ModuleWebsocketServer {
	mws := &ModuleWebsocketServer{}
	for _, mo := range mos {
		mo(mws)
	}
	return mws
}

func (*ModuleWebsocketServer) WithPort(p string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleWebsocketServer).port = p }
}

func (*ModuleWebsocketServer) WithRouters(r []*HttpRouteInfo) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleWebsocketServer).routers = r }
}

func (m *ModuleWebsocketServer) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted, http server listen on url with dial path", zap.Any("url", m.port))

	// 创建 gin 引擎
	m.engine = gin.New()
	m.engine.Use(gin.Recovery())

	for _, router := range m.routers {
		switch router.Method {
		case http.MethodGet:
			m.engine.GET(router.Path, func(ctx *gin.Context) {
				router.Handler(&WebsocketContext{
					ModuleWebsocketServer: m,
					Context:               ctx,
				})
			})
		case http.MethodPost:
			m.engine.POST(router.Path, func(ctx *gin.Context) {
				router.Handler(&WebsocketContext{
					ModuleWebsocketServer: m,
					Context:               ctx,
				})
			})
		}
	}

	// 创建 http 服务器
	m.server = &http.Server{
		Addr:    ":" + m.port,
		Handler: m.engine,
	}

	// 启动 http 服务器
	go func() {
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.Logger().Error("Mounted, ListenAndServe occurs error", zap.Error(err), zap.Any("url", m.port))
			// 无法开启监听不提供服务
			panic(err)
		}
	}()
}
