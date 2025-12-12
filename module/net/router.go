package moduleNet

type HttpRouter struct {
	routes      []*HttpRouteInfo
	middlewares []HttpHandler
}

func NewRouter(routes []*HttpRouteInfo, middlewares ...HttpHandler) *HttpRouter {
	return &HttpRouter{routes: routes, middlewares: middlewares}
}

func (r *HttpRouter) WithMiddleware(middlewares ...HttpHandler) *HttpRouter {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}
