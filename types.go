package ninja

import "net/http"

type path []string

type methods []string

type route struct {
	path    path
	methods methods
	handler http.Handler
}

type middlewireHandler func(http.ResponseWriter, *http.Request) bool

type middlewire struct {
	path    path
	methods methods
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
