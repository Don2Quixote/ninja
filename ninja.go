package ninjago

import (
	"net/http"
	"strings"
)

func (p path) matches(requestPath path) bool {
	if len(p) > len(requestPath) {
		return false
	}

	for i := range requestPath {
		if i >= len(p) {
			if p[len(p)-1] == "" {
				return true
			}
			return false
		} else if p[i] == "" {
			return true
		} else if requestPath[i] != p[i] {
			return false
		}
	}

	return true
}

func (m methods) has(method string) bool {
	if len(m) == 0 {
		return true
	}

	for _, allowedMethod := range m {
		if allowedMethod == method {
			return true
		}
	}

	return false
}

func CreateRouter(routesCount, middlewaresCount int) *Router {
	return &Router{
		routesMaxCount:      routesCount,
		middlewaresMaxCount: middlewaresCount,
		routes:              make([]route, 0, routesCount),
		middlewares:         make([]middleware, 0, middlewaresCount),
	}
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestPath := strings.Split(req.URL.Path, "/")[1:]

	for _, middleware := range r.middlewares {
		if !middleware.path.matches(requestPath) {
			continue
		}

		if !middleware.methods.has(req.Method) {
			continue
		}

		if middleware.async {
			go middleware.handler(res, req)
		} else {
			pass := middleware.handler(res, req)
			if !pass {
				return
			}
		}
	}

	for _, route := range r.routes {
		if !route.path.matches(requestPath) {
			continue
		}

		if !route.methods.has(req.Method) {
			continue
		}

		route.handler.ServeHTTP(res, req)
		break
	}
}

func ThroughMiddleware(handler func(http.ResponseWriter, *http.Request)) MiddlewareHandler {
	return func(res http.ResponseWriter, req *http.Request) bool {
		handler(res, req)
		return true
	}
}

func (r *Router) SetMiddleware(path string, handler MiddlewareHandler) *middleware {
	middleware := middleware{
		path:    strings.Split(path, "/")[1:],
		handler: handler,
	}

	index := len(r.middlewares)
	if index == r.middlewaresMaxCount {
		panic("Middlewares more than allowed max count")
	}

	r.middlewares = append(r.middlewares, middleware)

	return &r.middlewares[index]
}

func (m *middleware) Methods(methods ...string) *middleware {
	for _, method := range methods {
		m.methods = append(m.methods, method)
	}

	return m
}

// Undefined behavior if using request's data (working with http.ResponseWriter or *http.Request)
// Recommended only for job which doesn't require request's data
func (m *middleware) Async() *middleware {
	m.async = true
	return m
}

func (r *Router) HandleFunc(path string, handler func(http.ResponseWriter, *http.Request)) *route {
	route := route{
		path:    strings.Split(path, "/")[1:],
		handler: http.HandlerFunc(handler),
	}

	index := len(r.routes)
	if index == r.routesMaxCount {
		panic("Routes more than allowed max count")
	}

	r.routes = append(r.routes, route)

	return &r.routes[index]
}

func (r *Router) Handle(path string, handler http.Handler) *route {
	route := route{
		path:    strings.Split(path, "/")[1:],
		handler: handler,
	}

	index := len(r.routes)
	if index == r.routesMaxCount {
		panic("Routes more than allowed max count")
	}

	r.routes = append(r.routes, route)

	return &r.routes[index]
}

func (r *route) Methods(methods ...string) *route {
	for _, method := range methods {
		r.methods = append(r.methods, method)
	}

	return r
}
