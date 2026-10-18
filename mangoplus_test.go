package mangoplus_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestNewClient_Defaults(t *testing.T) {
	c, err := mangoplus.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if c.Registration == nil {
		t.Error("Registration service is nil")
	}
	if c.TitleList == nil {
		t.Error("TitleList service is nil")
	}
	if c.Title == nil {
		t.Error("Title service is nil")
	}
	if c.Manga == nil {
		t.Error("Manga service is nil")
	}
	if c.CookieJar() == nil {
		t.Error("CookieJar is nil")
	}
	if c.Secret() != "" {
		t.Errorf("Secret() = %q, want empty for unregistered client", c.Secret())
	}
}

func TestNewClient_OptionError(t *testing.T) {
	// WithHTTPClient(nil) returns an error that must bubble up.
	_, err := mangoplus.NewClient(mangoplus.WithHTTPClient(nil))
	if err == nil {
		t.Fatal("NewClient returned nil error for failing option, want non-nil")
	}
}

func TestNewClient_AutoRegisterPropagatesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	_, err := mangoplus.NewClient(
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
		mangoplus.WithAutoRegister(context.Background()),
	)
	if err == nil {
		t.Fatal("NewClient with failing auto-register returned nil error, want non-nil")
	}
	if !strings.Contains(err.Error(), "auto register") {
		t.Errorf("error does not mention auto register: %v", err)
	}
}

func TestClient_NewRequest_AbsoluteURL(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, err := c.NewRequest(context.Background(), http.MethodGet, "https://example.com/foo")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if req.URL.Host != "example.com" {
		t.Errorf("Host = %q, want example.com", req.URL.Host)
	}
	if req.URL.Path != "/foo" {
		t.Errorf("Path = %q, want /foo", req.URL.Path)
	}
}

func TestClient_NewRequest_RelativePath(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, err := c.NewRequest(context.Background(), http.MethodGet, "/manga_viewer")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if req.URL.Host == "" {
		t.Error("Host is empty, want base URL host")
	}
	if !strings.HasSuffix(req.URL.Path, "/manga_viewer") {
		t.Errorf("Path = %q, want to end with /manga_viewer", req.URL.Path)
	}
}

func TestClient_NewRequest_InvalidURL(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// A URL with a control byte should fail to parse.
	if _, err := c.NewRequest(context.Background(), http.MethodGet, "ht\x7ftp://bad"); err == nil {
		t.Error("NewRequest with malformed URL returned nil error, want non-nil")
	}
}

func TestClient_NewRequest_InvalidMethod(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.NewRequest(context.Background(), "bad method", "/foo"); err == nil {
		t.Error("NewRequest with invalid method returned nil error, want non-nil")
	}
}

func TestClient_NewRequest_SetsContentTypeOnBody(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req, err := c.NewRequest(context.Background(), http.MethodPost, "/foo",
		mangoplus.WithRequestBody("application/x-protobuf", strings.NewReader("data")))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if got := req.Header.Get("Content-Type"); got != "application/x-protobuf" {
		t.Errorf("Content-Type = %q, want application/x-protobuf", got)
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(b) != "data" {
		t.Errorf("body = %q, want data", b)
	}
}

func TestClient_Register_NoopWithExistingSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("register endpoint called despite existing secret")
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("existing"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	secret, err := c.Register(context.Background())
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if secret != "existing" {
		t.Errorf("Register returned %q, want existing", secret)
	}
}

func TestClient_Register_PropagatesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("gateway"))
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithAndroidID("deadbeef"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.Register(context.Background())
	if err == nil {
		t.Fatal("Register returned nil error, want *APIError")
	}
	var apiErr *mangoplus.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *mangoplus.APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadGateway)
	}
	if string(apiErr.RawBody) != "gateway" {
		t.Errorf("RawBody = %q, want gateway", apiErr.RawBody)
	}
}

func TestClient_Register_InvalidProtoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-proto"))
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithAndroidID("deadbeef"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.Register(context.Background()); err == nil {
		t.Error("Register returned nil error for invalid proto body, want non-nil")
	}
}

func TestClient_Register_ProtoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, errorResponse(&proto.ErrorResult{
			Action: proto.ErrorResult_UNAUTHORIZED,
			EnglishPopup: &proto.Popup_OSDefault{
				Subject: "Invalid user",
				Body:    "User is invalid.",
			},
		}))
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithAndroidID("deadbeef"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.Register(context.Background())
	if err == nil {
		t.Fatal("Register returned nil error, want *ProtoError")
	}
	var pErr *mangoplus.ProtoError
	if !errors.As(err, &pErr) {
		t.Fatalf("error type = %T, want *mangoplus.ProtoError", err)
	}
	if pErr.Action != mangoplus.ErrorActionUnauthorized {
		t.Errorf("Action = %q, want %q", pErr.Action, mangoplus.ErrorActionUnauthorized)
	}
	if !errors.Is(err, mangoplus.ErrInvalidUserAccess) {
		t.Errorf("errors.Is(err, ErrInvalidUserAccess) = false, want true")
	}
}
