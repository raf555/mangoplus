package mangoplus_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

// capturingHandler returns an [http.HandlerFunc] that records the last
// observed request into *got and responds with 200 and an empty success body.
func capturingHandler(t *testing.T, got **http.Request) http.HandlerFunc {
	t.Helper()

	emptySuccess := &proto.Response{
		Result: &proto.Response_Success{Success: &proto.SuccessResult{}},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		*got = r.Clone(context.Background())
		writeProtoResponse(t, w, emptySuccess)
	}
}

// doAPICall triggers an arbitrary service call through c so the request runs
// through the client's configured http.Client (and therefore its transport).
// The returned error is intentionally ignored — tests inspect the captured
// request, not the response.
func doAPICall(t *testing.T, c *mangoplus.Client) {
	t.Helper()

	_, _ = c.TitleList.AllV2(context.Background())
}

func TestWithHTTPClient_Nil(t *testing.T) {
	if _, err := mangoplus.NewClient(mangoplus.WithHTTPClient(nil)); err == nil {
		t.Fatal("NewClient(WithHTTPClient(nil)) returned nil error, want non-nil")
	}
}

func TestWithHTTPClient_CopiesClient(t *testing.T) {
	hc := &http.Client{Timeout: 1 * time.Second}
	c, err := mangoplus.NewClient(mangoplus.WithHTTPClient(hc), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Mutating the original http.Client after construction must not leak into
	// the already-constructed *Client. We observe this via the cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}
	hc.Jar = jar

	if c.CookieJar() == jar {
		t.Error("mutation of the external http.Client leaked into *Client")
	}
}

func TestWithHTTPClient_KeepsJar(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}

	c, err := mangoplus.NewClient(mangoplus.WithHTTPClient(&http.Client{Jar: jar}), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.CookieJar() != jar {
		t.Error("CookieJar() did not return the jar of the provided http.Client")
	}
}

func TestWithTransport_Nil(t *testing.T) {
	if _, err := mangoplus.NewClient(mangoplus.WithTransport(nil)); err == nil {
		t.Fatal("NewClient(WithTransport(nil)) returned nil error, want non-nil")
	}
}

func TestWithTransport_IsUsed(t *testing.T) {
	called := false
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	c, err := mangoplus.NewClient(mangoplus.WithTransport(rt), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Any service call forces the configured transport to execute.
	_, _ = c.TitleList.AllV2(context.Background())

	if !called {
		t.Error("configured RoundTripper was not invoked by any client API call")
	}
}

func TestWithTimeout_Negative(t *testing.T) {
	if _, err := mangoplus.NewClient(mangoplus.WithTimeout(-1 * time.Second)); err == nil {
		t.Fatal("NewClient(WithTimeout(-1s)) returned nil error, want non-nil")
	}
}

func TestWithTimeout_Zero(t *testing.T) {
	if _, err := mangoplus.NewClient(mangoplus.WithTimeout(0), mangoplus.WithSecret("s")); err != nil {
		t.Fatalf("NewClient(WithTimeout(0)): %v", err)
	}
}

func TestWithTimeout_Enforced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTimeout(50*time.Millisecond),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.TitleList.AllV2(context.Background()); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWithCookieJar_Nil(t *testing.T) {
	if _, err := mangoplus.NewClient(mangoplus.WithCookieJar(nil)); err == nil {
		t.Fatal("NewClient(WithCookieJar(nil)) returned nil error, want non-nil")
	}
}

func TestWithCookieJar_Used(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}

	c, err := mangoplus.NewClient(mangoplus.WithCookieJar(jar), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.CookieJar() != jar {
		t.Error("CookieJar() did not return the provided jar")
	}
}

func TestWithCookieJar_DefaultIsNonNil(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.CookieJar() == nil {
		t.Error("default CookieJar is nil, want non-nil")
	}
}

func TestWithUserAgent(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got), mangoplus.WithUserAgent("MyAgent/9.9"))

	doAPICall(t, c)

	if got == nil {
		t.Fatal("no request captured")
	}
	if ua := got.Header.Get("User-Agent"); ua != "MyAgent/9.9" {
		t.Errorf("User-Agent = %q, want MyAgent/9.9", ua)
	}
}

func TestWithUserAgent_Default(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got))

	doAPICall(t, c)
	if got == nil {
		t.Fatal("no request captured")
	}
	if ua := got.Header.Get("User-Agent"); ua == "" {
		t.Error("default User-Agent is empty")
	}
}

func TestWithOSVersion(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got), mangoplus.WithOSVersion("34"))

	doAPICall(t, c)
	if got == nil {
		t.Fatal("no request captured")
	}
	if v := got.URL.Query().Get("os_ver"); v != "34" {
		t.Errorf("os_ver = %q, want 34", v)
	}
}

func TestWithAppVersion(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got), mangoplus.WithAppVersion("777"))

	doAPICall(t, c)
	if got == nil {
		t.Fatal("no request captured")
	}
	if v := got.URL.Query().Get("app_ver"); v != "777" {
		t.Errorf("app_ver = %q, want 777", v)
	}
}

func TestRequest_SendsOSAndroid(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got))

	doAPICall(t, c)
	if got == nil {
		t.Fatal("no request captured")
	}
	if v := got.URL.Query().Get("os"); v != "android" {
		t.Errorf("os = %q, want android", v)
	}
}

func TestWithSecret_AppearsInQuery(t *testing.T) {
	var got *http.Request
	c, _ := newTestServer(t, capturingHandler(t, &got), mangoplus.WithSecret("sekret"))
	// newTestServer already sets a secret; override it by rebuilding the
	// request below. Instead we simply trust the override above and issue a
	// standard service call which adds the configured secret automatically.
	doAPICall(t, c)

	if got == nil {
		t.Fatal("no request captured")
	}
	// newTestServer's default secret is "test-secret"; the last secret option
	// wins. Guard against both having-set and the chosen value.
	if v := got.URL.Query().Get("secret"); v != "sekret" {
		t.Errorf("secret = %q, want sekret", v)
	}
}

func TestWithoutSecret_OmitsSecret(t *testing.T) {
	// WithoutSecret is a request option; we cannot stuff it through the
	// service helpers which always include the secret. Instead invoke the
	// registration flow: Registration.Register internally passes
	// WithoutSecret() to NewRequest, so the captured request should NOT
	// carry the secret query param.
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_RegisterationData{
						RegisterationData: &proto.RegistrationData{DeviceSecret: "x"},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithAndroidID("deadbeef"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Register(context.Background()); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if got == nil {
		t.Fatal("no request captured")
	}
	if got.URL.Query().Has("secret") {
		t.Errorf("secret should be omitted, got %q", got.URL.Query().Get("secret"))
	}
}

func TestWithAndroidIDAndSecret_AreMutuallyExclusive(t *testing.T) {
	_, err := mangoplus.NewClient(
		mangoplus.WithAndroidID("abcdef0123456789"),
		mangoplus.WithSecret("sekret"),
	)
	if err == nil {
		t.Fatal("NewClient with both AndroidID and Secret returned nil error, want non-nil")
	}
}

func TestWithAutoRegister_NilCtx(t *testing.T) {
	var ctx context.Context
	if _, err := mangoplus.NewClient(mangoplus.WithAutoRegister(ctx)); err == nil {
		t.Fatal("WithAutoRegister(nil) accepted, want error")
	}
}

func TestWithAutoRegister_InvokesRegister(t *testing.T) {
	var registerPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		registerPath = r.URL.Path
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_RegisterationData{
						RegisterationData: &proto.RegistrationData{DeviceSecret: "auto-secret"},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
		mangoplus.WithAutoRegister(context.Background()),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if registerPath != "/api/register" {
		t.Errorf("register path = %q, want /api/register", registerPath)
	}
	if got := c.Secret(); got != "auto-secret" {
		t.Errorf("Secret() = %q, want auto-secret", got)
	}
}

func TestWithAutoRegister_NoopWhenSecretProvided(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("existing"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
		mangoplus.WithAutoRegister(context.Background()),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if called {
		t.Error("register endpoint was called despite secret being provided")
	}
	if got := c.Secret(); got != "existing" {
		t.Errorf("Secret() = %q, want existing", got)
	}
}

func TestWithRequestBody(t *testing.T) {
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	req, err := c.NewRequest(context.Background(), http.MethodPost, "/anything",
		mangoplus.WithRequestBody("application/json", strings.NewReader("{}")))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if req.Body == nil {
		t.Error("Body is nil, want non-nil")
	}
}
