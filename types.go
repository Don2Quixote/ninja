package ninjago

import "net/http"

type path []string

type methods []string

type route struct {
	path    path
	methods methods
	handler http.Handler
}

type MiddlewareHandler func(http.ResponseWriter, *http.Request) bool

type middleware struct {
	path    path
	methods methods
	async   bool
	handler MiddlewareHandler
}

// The reason I used max count is avoiding using pointers ([]*route and []*Middleware)
type Router struct {
	routesMaxCount      int
	middlewaresMaxCount int
	routes              []route
	middlewares         []middleware
}
