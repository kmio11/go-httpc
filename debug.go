package httpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sync"
)

// PrintMiddleware returns a MiddlewareFunc that prints the specified
// strings before and after the HTTP request is processed by the next
// RoundTripper in the chain.
func PrintMiddleware(before, after string) MiddlewareFunc {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if before != "" {
				fmt.Print(before)
			}
			resp, err := next.RoundTrip(req)
			if after != "" {
				fmt.Print(after)
			}
			return resp, err
		})
	}
}

// DumpMiddleware returns a MiddlewareFunc that logs HTTP requests and responses
// to the provided io.Writer.
func DumpMiddleware(w io.Writer) MiddlewareFunc {
	var mu sync.Mutex
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			func() {
				mu.Lock()
				defer mu.Unlock()
				if b, err := httputil.DumpRequest(req, true); err == nil {
					fmt.Fprintf(w, "-----request:\n%s\n-----\n", string(b))
				}
				if b, err := httputil.DumpRequestOut(req, true); err == nil {
					fmt.Fprintf(w, "-----outgoing request:\n%s\n-----\n", string(b))
				}
			}()

			resp, err := next.RoundTrip(req)
			if err != nil {
				return resp, err
			}

			func() {
				mu.Lock()
				defer mu.Unlock()
				if b, err := httputil.DumpResponse(resp, true); err == nil {
					fmt.Fprintf(w, "-----response:\n%s\n-----\n", string(b))
				}
			}()

			return resp, err
		})
	}
}
