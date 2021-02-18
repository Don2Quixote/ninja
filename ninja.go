package ninja

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
			return true
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

	for _, method := range m {
		if method == method {
			return true
		}
	}

	return false
}

func CreateRouter(routesCount, middlewiresCount int) *Router {
	return &Router{
		routesMaxCount:      routesCount,
		middlewiresMaxCount: middlewiresCount,
		routes:              make([]route, 0, routesCount),
		middlewires:         make([]middlewire, 0, middlewiresCount),
	}
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestPath := strings.Split(req.URL.Path, "/")[1:]

	for _, middlewire := range r.middlewires {
		if !middlewire.path.matches(requestPath) {
			continue
		}

		if !middlewire.methods.has(req.Method) {
			continue
		}

		if middlewire.async {
			go middlewire.handler(res, req)
		} else {
			pass := middlewire.handler(res, req)
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

func ThroughMiddlewire(handler func(http.ResponseWriter, *http.Request)) middlewireHandler {
	return func(res http.ResponseWriter, req *http.Request) bool {
		handler(res, req)
		return true
	}
}

func (r *Router) SetMiddlewire(path string, handler middlewireHandler) *middlewire {
	middlewire := middlewire{
		path:    strings.Split(path, "/")[1:],
		handler: handler,
	}

	index := len(r.middlewires)
	if index == r.middlewiresMaxCount {
		panic("Middlewires more than allowed max count")
	}

	r.middlewires = append(r.middlewires, middlewire)

	return &r.middlewires[index]
}

func (m *middlewire) Methods(methods ...string) *middlewire {
	for _, method := range methods {
		m.methods = append(m.methods, method)
	}

	return m
}

// Undefined behavior if using request's data (working with http.ResponseWriter or *http.Request)
// Recommended only for job which doesn't require request's data
func (m *middlewire) Async() *middlewire {
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
