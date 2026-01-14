package middleware

import (
	"net/http"

	moduleNet "github.com/Mericusta/go-sgs/module/net"
	"go.uber.org/zap"
)

var (
	unmountToken = "whosyourdaddy"
)

// 主服务卸载标识鉴权中间件
func SingleServerRouterMiddleware(ctx moduleNet.IHttpContext) {
	var (
		requestPath = ctx.Raw().Request.URL.Path
		token       = ctx.Raw().Param("token")
	)

	if token != unmountToken {
		ctx.Module().Base().Logger().Info("SingleServerUnmountMarkMiddleware, invalid token, abort", zap.Any("token", token), zap.Any("requestPath", requestPath))
		ctx.Raw().AbortWithStatus(http.StatusInternalServerError)
		return
	}
}

var (
	healthCheckPaths = []string{"/health"}
	healthCheckAgent = "ELB-HealthChecker/2.0"
)

// 主服务 K8S 健康检查过滤中间件
func SingleServerK8SHealthCheckMiddleware(ctx moduleNet.IHttpContext) {
	var (
		requestPath   = ctx.Raw().Request.URL.Path
		requestHeader = ctx.Raw().Request.Header
		userAgent     = requestHeader.Get("User-Agent")
	)

	if userAgent == healthCheckAgent {
		for _, healthCheckPath := range healthCheckPaths {
			if requestPath == healthCheckPath {
				ctx.Raw().AbortWithStatus(http.StatusOK)
				ctx.Module().Base().Logger().Info("SingleServerK8SHealthCheckMiddleware, abort")
				return
			}
		}
	}

	ctx.Module().Base().Logger().Info("SingleServerK8SHealthCheckMiddleware, continue", zap.Any("userAgent", userAgent), zap.Any("requestPath", requestPath))
}
