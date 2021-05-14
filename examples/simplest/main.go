package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	ninja "github.com/don2quixote/ninjago"
)

func main() {
	// Creating router with setting the max value of handlers and middlewares we will pass to it
	router := ninja.CreateRouter(2, 5)

	// Making this middleware async because of no need in data access
	// Actually, I don't recommend to make such easy job to be async - it's harder to create goroutine than place a line in output
	router.SetMiddleware("", ninja.ThroughMiddleware(notifyAboutRequest)).Async()

	// Making these middlewares async is not safe because of danger concurrent access to data (variables-counters)
	// ThroughMiddleware passes each request (function-handler doesn't have to return bool value)
	// Empty path ("") matches to any path
	router.SetMiddleware("", ninja.ThroughMiddleware(countAllRequests))
	router.SetMiddleware("", ninja.ThroughMiddleware(countGetRequests)).Methods("GET")
	router.SetMiddleware("", ninja.ThroughMiddleware(countPostRequests)).Methods("POST")

	// If accessMiddleware function return false, next /access handler won't be called
	router.SetMiddleware("/access", accessMiddleware).Methods("GET")

	router.HandleFunc("/access", handleAccess).Methods("GET")

	// If no of handlers above (exclude middlewares) matched, this will match as it's path ("") matches any path
	// It also shows that order of handlers is important
	router.HandleFunc("", handleNotFound)

	err := http.ListenAndServe(":80", router)
	if err != nil {
		fmt.Println("Error launching server:", err)
	}
}

func notifyAboutRequest(res http.ResponseWriter, req *http.Request) {
	fmt.Println("New requests recieved")
}

var (
	allRequestsCount  int32 = 0
	getRequestsCount  int32 = 0
	postRequestsCount int32 = 0
)

func countAllRequests(res http.ResponseWriter, req *http.Request) {
	atomic.AddInt32(&allRequestsCount, 1)
	fmt.Printf("All requests count: %d\n", allRequestsCount)
}

func countGetRequests(res http.ResponseWriter, req *http.Request) {
	atomic.AddInt32(&getRequestsCount, 1)
	fmt.Printf("GET requests count: %d\n", getRequestsCount)
}

func countPostRequests(res http.ResponseWriter, req *http.Request) {
	atomic.AddInt32(&postRequestsCount, 1)
	fmt.Printf("POST requests count: %d\n", postRequestsCount)
}

func accessMiddleware(res http.ResponseWriter, req *http.Request) bool {
	pass := req.URL.Query().Get("pass")
	accessGranted := pass == "test"

	if !accessGranted {
		res.Write([]byte("Access denied"))
		// Returning false from middleware prevents calling of next middlewares and handlers
		return false
	}

	return true
}

func handleAccess(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Access granted"))
}

func handleNotFound(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Not found. Maybe you are looking for /access path?"))
}
