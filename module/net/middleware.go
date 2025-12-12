package moduleNet

type IMiddleware interface {
}

type middleware struct {
	f []func(IHttpContext)
	i int
}

func (m *middleware) hijack(ctx *HttpContext) bool {
	// _f := m.f[m.i]

	return m.next(ctx)
}

func (m *middleware) next(ctx *HttpContext) bool {
	m.i++
	return m.hijack(ctx)
}
