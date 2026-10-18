package mangoplus_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"

	protopkg "google.golang.org/protobuf/proto"
)

// roundTripperFunc adapts an ordinary function to [http.RoundTripper].
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// rewriteTransport returns a [http.RoundTripper] that forwards every request
// to serverURL, preserving the request's path and query. It lets a [*Client]
// configured with the production base URL be redirected at a mock
// [httptest.Server] without exposing a base URL option.
func rewriteTransport(t *testing.T, serverURL string) http.RoundTripper {
	t.Helper()

	u, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("url.Parse(%q): %v", serverURL, err)
	}

	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		return http.DefaultTransport.RoundTrip(req)
	})
}

// newTestServer spins up an [httptest.Server] running handler and returns a
// [*mangoplus.Client] configured to send requests to it. The server is closed
// via [t.Cleanup].
func newTestServer(t *testing.T, handler http.HandlerFunc, opts ...mangoplus.ClientOptionsFunc) (*mangoplus.Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	all := []mangoplus.ClientOptionsFunc{
		mangoplus.WithSecret("test-secret"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	}
	all = append(all, opts...)

	c, err := mangoplus.NewClient(all...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

// marshalResponse encodes resp to the protobuf wire format or fails t.
func marshalResponse(t *testing.T, resp *proto.Response) []byte {
	t.Helper()

	b, err := protopkg.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	return b
}

// writeProtoResponse marshals resp and writes it to w with status 200.
func writeProtoResponse(t *testing.T, w http.ResponseWriter, resp *proto.Response) {
	t.Helper()
	if _, err := w.Write(marshalResponse(t, resp)); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

// errorResponse wraps err in a [*proto.Response] of error type.
func errorResponse(err *proto.ErrorResult) *proto.Response {
	return &proto.Response{Result: &proto.Response_Error{Error: err}}
}
