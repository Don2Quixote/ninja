package ninja

import (
	"net/http"
	"strings"
)

type route struct {
	path    []string
	methods []string
	handler http.Handler
}

type middlewireHandler func(http.ResponseWriter, *http.Request) bool

type middlewire struct {
	path    []string
	methods []string
	async   bool
	handler middlewireHandler
}

// The reason I used max count is avoiding using pointers ([]*route and []*Middlewire)
type Router struct {
	routesMaxCount      int
	middlewiresMaxCount int
	routes              []route
	middlewires         []middlewire
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
		pathMatches := true
		for i := range requestPath {
			if len(middlewire.path) <= i {
				break
			} else if middlewire.path[i] == "" {
				break
			} else if requestPath[i] != middlewire.path[i] {
				pathMatches = false
				break
			}
		}

		if !pathMatches {
			continue
		}

		methodMatches := false
		if len(middlewire.methods) == 0 {
			methodMatches = true
		} else {
			for _, method := range middlewire.methods {
				if method == req.Method {
					methodMatches = true
					break
				}
			}
		}

		if !methodMatches {
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
		pathMatches := true
		for i := range requestPath {
			if len(route.path) <= i {
				break
			} else if route.path[i] == "" {
				break
			} else if requestPath[i] != route.path[i] {
				pathMatches = false
				break
			}
		}

		if !pathMatches {
			continue
		}

		methodMatches := false
		if len(route.methods) == 0 {
			methodMatches = true
		} else {
			for _, method := range route.methods {
				if method == req.Method {
					methodMatches = true
				}
			}
		}

		if !methodMatches {
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
