package ninjago_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/don2quixote/ninjago"
)

func TestServeHTTP(t *testing.T) {
	router1 := CreateRouter(1, 0)
	router1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	router2 := CreateRouter(3, 0)
	router2.HandleFunc("/1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1"))
	})
	router2.HandleFunc("/2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
	})
	router2.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("any"))
	})

	router3 := CreateRouter(1, 1)
	router3.SetMiddleware("/test", func(w http.ResponseWriter, r *http.Request) bool {
		value := r.URL.Query().Get("value")
		if value == "" {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		return true
	})
	router3.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router4 := CreateRouter(3, 0)
	router4.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("POST"))
	}).Methods("POST")
	router4.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GET"))
	}).Methods("GET")
	router4.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	router5 := CreateRouter(2, 0)
	router5.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Split(r.URL.Path, "/")
		if path[len(path)-1] == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Write([]byte(path[len(path)-1]))
	})
	router5.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	type testingReq struct {
		path         string
		method       string
		code         int
		checkBody    bool
		expectedBody string
	}
	testCases := []struct {
		router *Router
		reqs   []testingReq
	}{
		{router1, []testingReq{
			{"/", "GET", http.StatusOK, true, "test"},
			{"/test", "GET", http.StatusOK, true, "test"},
			{"/test/", "GET", http.StatusOK, true, "test"},
			{"/test/path", "GET", http.StatusOK, true, "test"},
		}},
		{router2, []testingReq{
			{"/", "GET", http.StatusOK, true, "any"},
			{"/1", "GET", http.StatusOK, true, "1"},
			{"/2", "GET", http.StatusOK, true, "2"},
		}},
		{router3, []testingReq{
			{"/test", "GET", http.StatusForbidden, false, ""},
			{"/test?value=test", "GET", http.StatusOK, false, ""},
		}},
		{router4, []testingReq{
			{"/", "GET", http.StatusOK, true, "GET"},
			{"/", "POST", http.StatusOK, true, "POST"},
			{"/", "PUT", http.StatusNotFound, false, ""},
		}},
		{router5, []testingReq{
			{"/", "GET", http.StatusNotFound, false, ""},
			{"/test", "GET", http.StatusNotFound, false, ""},
			{"/test/", "GET", http.StatusForbidden, false, ""},
			{"/test/hello", "GET", http.StatusOK, true, "hello"},
			{"/test/kinkong/godzilla", "GET", http.StatusOK, true, "godzilla"},
		}},
	}

	for _, test := range testCases {
		for _, req := range test.reqs {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(req.method, fmt.Sprintf("http://localhost%s", req.path), nil)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}
			test.router.ServeHTTP(w, r)
			if w.Result().StatusCode != req.code {
				t.Fatalf("Expected code %d on request %s %s but got code %d", req.code, req.method, req.path, w.Result().StatusCode)
			}
			if req.checkBody {
				body, err := ioutil.ReadAll(w.Result().Body)
				if err != nil {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
				if string(body) != req.expectedBody {
					t.Fatalf("Expected body: %s, Got: %s", req.expectedBody, body)
				}
			}
		}
	}
}
