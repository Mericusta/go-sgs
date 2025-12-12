package moduleNet

import (
	"errors"
	"net/http"

	"github.com/Mericusta/go-sgs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ModuleHttpServer struct {
	sgs.ModuleBase

	port        string
	routers     []*HttpRouter
	middlewares []HttpHandler // global middlewares

	engine *gin.Engine
	server *http.Server
}

var HttpServer *ModuleHttpServer

func (*ModuleHttpServer) New(mos ...sgs.ModuleOption) *ModuleHttpServer {
	mws := &ModuleHttpServer{}
	for _, mo := range mos {
		mo(mws)
	}
	return mws
}

func (*ModuleHttpServer) WithPort(p string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleHttpServer).port = p }
}

func (*ModuleHttpServer) WithRouters(r ...*HttpRouter) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleHttpServer).routers = r }
}

func (*ModuleHttpServer) WithMiddleware(f HttpHandler) sgs.ModuleOption {
	return func(m sgs.Module) {
		m.(*ModuleHttpServer).middlewares = append(m.(*ModuleHttpServer).middlewares, f)
	}
}

func (m *ModuleHttpServer) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted, http server listen on url with dial path", zap.Any("url", m.port))

	m.engine = gin.New()
	m.engine.Use(gin.Recovery())
	// global middleware
	for _, middleware := range m.middlewares {
		m.engine.Use(func(ctx *gin.Context) {
			middleware(&HttpContext{ModuleHttpServer: m, Context: ctx})
		})
	}

	for _, router := range m.routers {
		for _, route := range router.routes {
			_middlewares := make([]gin.HandlerFunc, 0, 8)
			for _, middleware := range router.middlewares {
				_middlewares = append(_middlewares, func(ctx *gin.Context) { middleware(&HttpContext{ModuleHttpServer: m, Context: ctx}) })
			}
			_middlewares = append(_middlewares, func(ctx *gin.Context) { route.Handler(&HttpContext{ModuleHttpServer: m, Context: ctx}) })
			switch route.Method {
			case http.MethodGet:
				m.engine.GET(route.Path, _middlewares...)
			case http.MethodPost:
				m.engine.POST(route.Path, _middlewares...)
			}
		}
	}

	m.server = &http.Server{
		Addr:    ":" + m.port,
		Handler: m.engine,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.Logger().Error("http server listen on url occurs error", zap.Error(err), zap.Any("url", m.port))
			panic(err)
		}
	}()
}

func (m *ModuleHttpServer) handleMiddlewares(ctx *gin.Context) {

}
