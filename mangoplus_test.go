package mangoplus_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"

	protopkg "google.golang.org/protobuf/proto"
)

// baseURLPath is a non-empty Client base URL path to use during tests,
// to ensure relative URLs are used for all endpoints.
const baseURLPath = "/api"

// setup sets up a test HTTP server along with a Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
//
// Options that send requests during NewClient (e.g. WithAutoRegister) can't
// be used here, since the client only points at the test server afterwards.
func setup(t *testing.T, opts ...mangoplus.ClientOptionsFunc) (client *mangoplus.Client, mux *http.ServeMux, serverURL string) {
	t.Helper()

	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("request path %q does not start with Client base URL path %q", r.URL.Path, baseURLPath)
		http.Error(w, "Client base URL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)
	t.Cleanup(server.Close)

	client, err := mangoplus.NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	u, err := url.Parse(server.URL + baseURLPath)
	if err != nil {
		t.Fatalf("url.Parse returned error: %v", err)
	}
	mangoplus.SetBaseURL(client, u)

	return client, mux, server.URL
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

// testFormValues checks that the request query matches values exactly.
func testFormValues(t *testing.T, r *http.Request, values values) {
	t.Helper()
	want := url.Values{}
	for k, v := range values {
		want.Set(k, v)
	}

	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm returned error: %v", err)
	}
	if got := r.Form; !cmp.Equal(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

// testWriteResponse marshals res and writes it to w. It runs inside handlers,
// so it reports failures with t.Errorf rather than t.Fatalf.
func testWriteResponse(t *testing.T, w http.ResponseWriter, res *proto.Response) {
	t.Helper()
	b, err := protopkg.Marshal(res)
	if err != nil {
		t.Errorf("proto.Marshal returned error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(b); err != nil {
		t.Errorf("Write returned error: %v", err)
	}
}

func testWriteSuccess(t *testing.T, w http.ResponseWriter, res *proto.SuccessResult) {
	t.Helper()
	testWriteResponse(t, w, &proto.Response{Result: &proto.Response_Success{Success: res}})
}

func testWriteError(t *testing.T, w http.ResponseWriter, res *proto.ErrorResult) {
	t.Helper()
	testWriteResponse(t, w, &proto.Response{Result: &proto.Response_Error{Error: res}})
}

// testNewRequestAndDoFailure checks that f, which calls the API method under
// test with the given ctx and client, returns an error when the request can't
// be built and when it can't be sent.
func testNewRequestAndDoFailure(t *testing.T, methodName string, f func(ctx context.Context, client *mangoplus.Client) error) {
	t.Helper()
	if methodName == "" {
		t.Error("testNewRequestAndDoFailure: must supply method methodName")
	}

	client, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("transport failure")
		})),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	// http.NewRequestWithContext rejects a nil context.
	var nilCtx context.Context
	if err := f(nilCtx, client); err == nil {
		t.Errorf("nil context: %v err = nil, want error", methodName)
	}

	if err := f(t.Context(), client); err == nil {
		t.Errorf("failing transport: %v err = nil, want error", methodName)
	}
}

// testRequest builds a client through setup with opts, sends a request with
// it and returns the request as seen by the server.
func testRequest(t *testing.T, opts ...mangoplus.ClientOptionsFunc) *http.Request {
	t.Helper()
	client, mux, _ := setup(t, opts...)

	var got *http.Request
	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(r.Context())
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_AllTitlesViewV2{AllTitlesViewV2: &proto.AllTitlesViewV2{}},
		})
	})

	if _, err := client.TitleList.AllV2(t.Context()); err != nil {
		t.Fatalf("TitleList.AllV2 returned error: %v", err)
	}
	if got == nil {
		t.Fatal("no request received by the server")
	}
	return got
}

// roundTripperFunc adapts an ordinary function to [http.RoundTripper].
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// rewriteTransport returns a [http.RoundTripper] that forwards every request
// to serverURL, preserving the request's path and query. It's only needed for
// options that send requests during NewClient, before setup can point the
// client at the test server.
func rewriteTransport(t *testing.T, serverURL string) http.RoundTripper {
	t.Helper()

	u, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("url.Parse(%q) returned error: %v", serverURL, err)
	}

	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		return http.DefaultTransport.RoundTrip(req)
	})
}

func TestNewClient_Success(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if c.Registration == nil || c.TitleList == nil || c.Title == nil || c.Manga == nil {
		t.Error("NewClient left a service nil")
	}
	if c.Secret() != "" {
		t.Errorf("NewClient Secret() is %q, want empty for unregistered client", c.Secret())
	}
	if c.CookieJar() == nil {
		t.Error("NewClient CookieJar() is nil")
	}

	req, err := c.NewRequest(t.Context(), http.MethodGet, "/foo")
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	want := mangoplus.AppAPIURL + "/foo?app_ver=" + mangoplus.DefaultAppVersion + "&os=android&os_ver=" + mangoplus.DefaultOSVersion
	if got := req.URL.String(); got != want {
		t.Errorf("NewClient request URL is %v, want %v", got, want)
	}
	testHeader(t, req, "User-Agent", mangoplus.DefaultUserAgent)
}

func TestNewClient_OptionError(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithHTTPClient(nil)); err == nil {
		t.Error("NewClient returned nil error for failing option, want error")
	}
}

func TestNewClient_AutoRegisterError(t *testing.T) {
	t.Parallel()
	_, mux, serverURL := setup(t)

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := mangoplus.NewClient(
		mangoplus.WithTransport(rewriteTransport(t, serverURL)),
		mangoplus.WithAutoRegister(t.Context()),
	)
	if err == nil {
		t.Fatal("NewClient returned nil error for failing auto register, want error")
	}
	if !strings.Contains(err.Error(), "auto register") {
		t.Errorf("NewClient error %q does not mention auto register", err)
	}
}

func TestNewRequest_Success(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"), mangoplus.WithUserAgent("ua"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := c.NewRequest(t.Context(), http.MethodGet, "/manga_viewer?chapter_id=1")
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	// test that relative URL was expanded and query params were attached
	want := mangoplus.AppAPIURL + "/manga_viewer?app_ver=" + mangoplus.DefaultAppVersion + "&chapter_id=1&os=android&os_ver=" + mangoplus.DefaultOSVersion + "&secret=s"
	if got := req.URL.String(); got != want {
		t.Errorf("NewRequest URL is %v, want %v", got, want)
	}

	// test that user-agent is attached to the request
	testHeader(t, req, "User-Agent", "ua")
}

func TestNewRequest_AbsoluteURL(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := c.NewRequest(t.Context(), http.MethodGet, "https://example.com/foo")
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if got, want := req.URL.Host+req.URL.Path, "example.com/foo"; got != want {
		t.Errorf("NewRequest URL is %v, want %v", got, want)
	}
}

func TestNewRequest_BadURL(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	// A URL with a control byte fails to parse.
	if _, err := c.NewRequest(t.Context(), http.MethodGet, "ht\x7ftp://bad"); err == nil {
		t.Error("NewRequest returned nil error for malformed URL, want error")
	}
}

func TestNewRequest_BadMethod(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	if _, err := c.NewRequest(t.Context(), "bad method", "/foo"); err == nil {
		t.Error("NewRequest returned nil error for invalid method, want error")
	}
}

func TestNewRequest_WithRequestBody(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := c.NewRequest(t.Context(), http.MethodPost, "/foo",
		mangoplus.WithRequestBody("application/x-protobuf", strings.NewReader("data")))
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	testHeader(t, req, "Content-Type", "application/x-protobuf")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if got, want := string(body), "data"; got != want {
		t.Errorf("NewRequest Body is %v, want %v", got, want)
	}
}

func TestNewRequest_WithoutSecret(t *testing.T) {
	t.Parallel()
	c, err := mangoplus.NewClient(mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	req, err := c.NewRequest(t.Context(), http.MethodGet, "/foo", mangoplus.WithoutSecret())
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if req.URL.Query().Has("secret") {
		t.Errorf("NewRequest URL %v contains secret, want it omitted", req.URL)
	}
}

func TestDo_HTTPError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("gateway"))
	})

	_, err := client.TitleList.AllV2(t.Context())

	var apiErr *mangoplus.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected *APIError, got %T: %v", err, err)
	}
	if got, want := apiErr.StatusCode, http.StatusBadGateway; got != want {
		t.Errorf("APIError.StatusCode is %v, want %v", got, want)
	}
	if got, want := string(apiErr.RawBody), "gateway"; got != want {
		t.Errorf("APIError.RawBody is %v, want %v", got, want)
	}
}

func TestDo_InvalidProtoBody(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-proto"))
	})

	if _, err := client.TitleList.AllV2(t.Context()); err == nil {
		t.Error("Expected error for invalid proto body, got nil")
	}
}

func TestDo_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			Action:       proto.ErrorResult_MAINTENANCE,
			EnglishPopup: &proto.Popup_OSDefault{Subject: "NEW VERSION AVAILABLE", Body: "update app", Language: proto.Language_ENGLISH},
			DebugInfo:    "debug",
			Popups: []*proto.Popup_OSDefault{
				{Subject: "NUEVA VERSIÓN", Body: "actualizar", Language: proto.Language_SPANISH},
			},
		})
	})

	_, err := client.TitleList.AllV2(t.Context())

	var pErr *mangoplus.ProtoError
	if !errors.As(err, &pErr) {
		t.Fatalf("Expected *ProtoError, got %T: %v", err, err)
	}
	want := &mangoplus.ProtoError{
		Action:       mangoplus.ErrorActionMaintenance,
		EnglishPopup: mangoplus.PopupOsDefault{Subject: "NEW VERSION AVAILABLE", Body: "update app", Language: mangoplus.LanguageEnglish},
		DebugInfo:    "debug",
		Popups: []mangoplus.PopupOsDefault{
			{Subject: "NUEVA VERSIÓN", Body: "actualizar", Language: mangoplus.LanguageSpanish},
		},
	}
	if diff := cmp.Diff(want, pErr); diff != "" {
		t.Errorf("ProtoError mismatch (-want +got):\n%s", diff)
	}
	if !errors.Is(err, mangoplus.ErrNewVersionAvailable) {
		t.Errorf("errors.Is(err, ErrNewVersionAvailable) = false, want true")
	}
}

func TestDo_ProtoErrorAction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pb   proto.ErrorResult_Action
		want mangoplus.ErrorAction
	}{
		{proto.ErrorResult_DEFAULT, mangoplus.ErrorActionDefault},
		{proto.ErrorResult_UNAUTHORIZED, mangoplus.ErrorActionUnauthorized},
		{proto.ErrorResult_MAINTENANCE, mangoplus.ErrorActionMaintenance},
		{proto.ErrorResult_GEOIP_BLOCKING, mangoplus.ErrorActionGeoIPBlocking},
		{proto.ErrorResult_Action(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.pb.String(), func(t *testing.T) {
			t.Parallel()
			client, mux, _ := setup(t)

			mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
				testWriteError(t, w, &proto.ErrorResult{Action: tt.pb})
			})

			_, err := client.TitleList.AllV2(t.Context())

			var pErr *mangoplus.ProtoError
			if !errors.As(err, &pErr) {
				t.Fatalf("Expected *ProtoError, got %T: %v", err, err)
			}
			if pErr.Action != tt.want {
				t.Errorf("ProtoError.Action is %q, want %q", pErr.Action, tt.want)
			}
		})
	}
}

func TestClient_Register_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithAndroidID("deadbeef"))

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testFormValues(t, r, values{
			"device_token": "4f41243847da693a4f356c0486114bc6", // md5("deadbeef")
			"security_key": "7695073a702ae7324213f9b66f4dc3fd", // md5(device_token + "4Kin9vGg")
			"os":           "android",
			"os_ver":       mangoplus.DefaultOSVersion,
			"app_ver":      mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_RegisterationData{
				RegisterationData: &proto.RegistrationData{DeviceSecret: "secret-123"},
			},
		})
	})

	secret, err := client.Register(t.Context())
	if err != nil {
		t.Fatalf("Client.Register returned error: %v", err)
	}
	if want := "secret-123"; secret != want {
		t.Errorf("Client.Register returned %v, want %v", secret, want)
	}
	if got, want := client.Secret(), "secret-123"; got != want {
		t.Errorf("Client.Secret() is %v, want %v", got, want)
	}
}

func TestClient_Register_ExistingSecret(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("existing"))

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		t.Error("register endpoint called despite existing secret")
	})

	secret, err := client.Register(t.Context())
	if err != nil {
		t.Fatalf("Client.Register returned error: %v", err)
	}
	if want := "existing"; secret != want {
		t.Errorf("Client.Register returned %v, want %v", secret, want)
	}
}

func TestClient_Register_Error(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithAndroidID("deadbeef"))

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			Action:       proto.ErrorResult_UNAUTHORIZED,
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Invalid user", Body: "User is invalid."},
		})
	})

	_, err := client.Register(t.Context())
	if !errors.Is(err, mangoplus.ErrInvalidUserAccess) {
		t.Errorf("Client.Register error = %v, want ErrInvalidUserAccess", err)
	}
	if client.Secret() != "" {
		t.Errorf("Client.Secret() is %q after failed register, want empty", client.Secret())
	}
}
