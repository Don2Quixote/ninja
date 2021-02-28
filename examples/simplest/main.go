package main

import (
	"fmt"
	"net/http"

	ninja "github.com/don2quixote/ninjago"
)

func main() {
	// Creating router with setting the max value of handlers and middlewires we will pass to it
	router := ninja.CreateRouter(2, 5)

	// Making this middlewire async because of no need in data access
	// Actually, I don't recommend to make such easy job to be async - it's harder to create goroutine than place a line in output
	router.SetMiddlewire("", ninja.ThroughMiddlewire(notifyAboutRequest)).Async()

	// Making these middlewires async is not safe because of danger concurrent access to data (variables-counters)
	// ThroughMiddlewire passes each request (function-handler doesn't have to return bool value)
	// Empty path ("") matches to any path
	router.SetMiddlewire("", ninja.ThroughMiddlewire(countAllRequests))
	router.SetMiddlewire("", ninja.ThroughMiddlewire(countGetRequests)).Methods("GET")
	router.SetMiddlewire("", ninja.ThroughMiddlewire(countPostRequests)).Methods("POST")

	// If accessMiddlewire function return false, next /access handler won't be called
	router.SetMiddlewire("/access", accessMiddlewire).Methods("GET")

	router.HandleFunc("/access", handleAccess).Methods("GET")

	// If no of handlers above (exclude middlewires) matched, this will match as it's path ("") matches any path
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
	allRequestsCount  = 0
	getRequestsCount  = 0
	postRequestsCount = 0
)

func countAllRequests(res http.ResponseWriter, req *http.Request) {
	allRequestsCount++
	fmt.Printf("All requests count: %d\n", allRequestsCount)
}

func countGetRequests(res http.ResponseWriter, req *http.Request) {
	getRequestsCount++
	fmt.Printf("GET requests count: %d\n", getRequestsCount)
}

func countPostRequests(res http.ResponseWriter, req *http.Request) {
	postRequestsCount++
	fmt.Printf("POST requests count: %d\n", postRequestsCount)
}

func accessMiddlewire(res http.ResponseWriter, req *http.Request) bool {
	pass := req.URL.Query().Get("pass")
	accessGranted := pass == "test"

	if !accessGranted {
		res.Write([]byte("Access denied"))
		// Returning false from middlewire prevents calling of next middlewires and handlers
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
