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

func TestCreateRouterWithRoutesOverflow1(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Fatal("Expected to panic due to routes overflow")
		}
	}()
	router := CreateRouter(0, 0)
	router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {})
}

func TestCreateRouterWithRoutesOverflow2(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Fatal("Expected to panic due to routes overflow")
		}
	}()
	router := CreateRouter(0, 0)
	router.Handle("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

func TestCreateRouterWithMiddlewaresOverflow(t *testing.T) {
	defer func() {
		if p := recover(); p == nil {
			t.Fatal("Expected to panic due to middlewares overflow")
		}
	}()
	router := CreateRouter(0, 0)
	router.SetMiddleware("", ThroughMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
}

func TestCreateRouterWithNoOverflows(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatal("Not expected to panic - no overflows")
		}
	}()
	router := CreateRouter(1, 1)
	router.SetMiddleware("", ThroughMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	router.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {})
}

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
	}).Methods("GET")
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

	router6 := CreateRouter(2, 1)
	router6.SetMiddleware("//", ThroughMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Could do some smart sruff here
		}
	})).Async()
	router6.Handle("//", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret place"))
	}))
	router6.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not secret"))
	})

	router7 := CreateRouter(3, 0)
	router7.HandleFunc("/working", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router7.HandleFunc("working/no1stslash", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router7.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	router8 := CreateRouter(3, 0)
	router8.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router8.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
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
			{"/test", "POST", http.StatusOK, false, ""},
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
		{router6, []testingReq{
			{"/", "GET", http.StatusOK, true, "not secret"},
			{"//", "POST", http.StatusOK, true, "secret place"},
		}},
		{router7, []testingReq{
			{"/working", "GET", http.StatusOK, false, ""},
			{"/working/no1stslash", "GET", http.StatusOK, false, ""},
			{"/", "GET", http.StatusNotFound, false, ""},
		}},
		{router8, []testingReq{
			{"/test/", "GET", http.StatusOK, false, ""},
			{"/test/mustwork", "GET", http.StatusOK, false, ""},
		}},
	}

	for _, test := range testCases {
		for _, req := range test.reqs {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(req.method, fmt.Sprintf("http://127.0.0.1%s", req.path), nil)
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
