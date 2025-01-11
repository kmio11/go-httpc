package httpc_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-redis/cache/v9"
	"github.com/kmio11/go-httpc"
	"github.com/kmio11/go-httpc/cachemw"
	"github.com/kmio11/go-httpc/cachemw/badgercache"
	"github.com/kmio11/go-httpc/cachemw/rediscache"
	"github.com/kmio11/go-httpc/cachemw/textcache"
	"github.com/redis/go-redis/v9"
)

// NewServer creates and returns a new httptest.Server for testing.
func NewServer() (testServer *httptest.Server, reset func()) {
	m := sync.RWMutex{}
	cnt := 0
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Lock()
		defer m.Unlock()
		fmt.Fprintf(w, "Hello, %d", cnt)
		cnt++
	})

	reset = func() {
		m.Lock()
		defer m.Unlock()
		cnt = 0
	}

	testServer = httptest.NewServer(h)
	return testServer, reset
}

func Example_rediscache() {
	ts, _ := NewServer()
	defer ts.Close()

	// a redis server for testing
	rs, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer rs.Close()

	rCache := cache.New(&cache.Options{
		Redis: redis.NewClient(&redis.Options{
			Addr: rs.Addr(),
			DB:   0,
		}),
	})

	// Create a new http client with a cache middleware.
	client := &http.Client{
		Transport: httpc.NewTransport().
			Use(
				httpc.PrintMiddleware("start\n", ""),
				cachemw.New(
					rediscache.New(rCache),
					cachemw.WithTTL(5*time.Minute),
				),
				httpc.PrintMiddleware("cache miss\n", "response from server\n"),
			),
	}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL, nil)
		if err != nil {
			panic(err)
		}

		fmt.Println("Request:", i)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Response: %s\n", string(respBody))
		fmt.Println("---")
	}

	// Output:
	// Request: 0
	// start
	// cache miss
	// response from server
	// Response: Hello, 0
	// ---
	// Request: 1
	// start
	// Response: Hello, 0
	// ---
	// Request: 2
	// start
	// Response: Hello, 0
	// ---
}

func Example_badgercache() {
	ts, _ := NewServer()
	defer ts.Close()

	// In-memory badger db
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a new http client with a cache middleware.
	client := &http.Client{
		Transport: httpc.NewTransport().
			Use(
				httpc.PrintMiddleware("start\n", ""),
				cachemw.New(
					badgercache.New(db),
					cachemw.WithTTL(5*time.Minute),
				),
				httpc.PrintMiddleware("cache miss\n", "response from server\n"),
			),
	}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL, nil)
		if err != nil {
			panic(err)
		}

		fmt.Println("Request:", i)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Response: %s\n", string(respBody))
		fmt.Println("---")
	}

	// Output:
	// Request: 0
	// start
	// cache miss
	// response from server
	// Response: Hello, 0
	// ---
	// Request: 1
	// start
	// Response: Hello, 0
	// ---
	// Request: 2
	// start
	// Response: Hello, 0
	// ---
}

func Example_textcache() {
	ts, _ := NewServer()
	defer ts.Close()

	// Create a temporary directory for the cache.
	cacheBaseDir := filepath.Join("testdata", "Example_textcache", "tmp", time.Now().Format("20060102150405.999"))

	// Create a new http client with a cache middleware.
	client := &http.Client{
		Transport: httpc.NewTransport().
			Use(
				httpc.PrintMiddleware("start\n", ""),
				cachemw.New(
					textcache.New(
						cacheBaseDir,
						textcache.WithDumpRequest(),
					),
				),
				httpc.PrintMiddleware("cache miss\n", "response from server\n"),
			),
	}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL, bytes.NewBufferString("Hello, World!"))
		if err != nil {
			panic(err)
		}

		fmt.Println("Request:", i)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Response: %s\n", string(respBody))
		fmt.Println("---")
	}

	// Output:
	// Request: 0
	// start
	// cache miss
	// response from server
	// Response: Hello, 0
	// ---
	// Request: 1
	// start
	// Response: Hello, 0
	// ---
	// Request: 2
	// start
	// Response: Hello, 0
	// ---
}
