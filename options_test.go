package mangoplus_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestWithHTTPClient_Success(t *testing.T) {
	t.Parallel()
	hc := &http.Client{}
	c, err := mangoplus.NewClient(mangoplus.WithHTTPClient(hc), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	// The provided http.Client is copied, so mutating it afterwards must not
	// leak into the Client.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	hc.Jar = jar
	if c.CookieJar() == jar {
		t.Error("mutation of the provided http.Client leaked into Client")
	}
}

func TestWithHTTPClient_Nil(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithHTTPClient(nil)); err == nil {
		t.Error("NewClient(WithHTTPClient(nil)) returned nil error, want error")
	}
}

func TestWithHTTPClient_KeepsJar(t *testing.T) {
	t.Parallel()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}

	c, err := mangoplus.NewClient(mangoplus.WithHTTPClient(&http.Client{Jar: jar}), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if c.CookieJar() != jar {
		t.Error("CookieJar() did not return the jar of the provided http.Client")
	}
}

func TestWithHTTPClient_Timeout(t *testing.T) {
	t.Parallel()
	hc := &http.Client{Timeout: 50 * time.Millisecond}
	client, mux, _ := setup(t, mangoplus.WithHTTPClient(hc), mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		// Block until the client gives up.
		<-r.Context().Done()
	})

	if _, err := client.TitleList.AllV2(t.Context()); err == nil {
		t.Error("TitleList.AllV2 returned nil error, want timeout error from the provided http.Client")
	}
}

func TestWithTransport_Success(t *testing.T) {
	t.Parallel()
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

	client, _, _ := setup(t, mangoplus.WithTransport(rt), mangoplus.WithSecret("s"))

	// The response is empty, so only the transport call matters here.
	_, _ = client.TitleList.AllV2(t.Context())

	if !called {
		t.Error("configured RoundTripper was not invoked")
	}
}

func TestWithTransport_Nil(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithTransport(nil)); err == nil {
		t.Error("NewClient(WithTransport(nil)) returned nil error, want error")
	}
}

func TestWithTimeout_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"), mangoplus.WithTimeout(50*time.Millisecond))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		// Block until the client gives up.
		<-r.Context().Done()
	})

	if _, err := client.TitleList.AllV2(t.Context()); err == nil {
		t.Error("TitleList.AllV2 returned nil error, want timeout error")
	}
}

func TestWithTimeout_Zero(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithTimeout(0), mangoplus.WithSecret("s")); err != nil {
		t.Errorf("NewClient(WithTimeout(0)) returned error: %v", err)
	}
}

func TestWithTimeout_Negative(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithTimeout(-time.Second)); err == nil {
		t.Error("NewClient(WithTimeout(-1s)) returned nil error, want error")
	}
}

func TestWithCookieJar_Success(t *testing.T) {
	t.Parallel()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}

	c, err := mangoplus.NewClient(mangoplus.WithCookieJar(jar), mangoplus.WithSecret("s"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if c.CookieJar() != jar {
		t.Error("CookieJar() did not return the provided jar")
	}
}

func TestWithCookieJar_Nil(t *testing.T) {
	t.Parallel()
	if _, err := mangoplus.NewClient(mangoplus.WithCookieJar(nil)); err == nil {
		t.Error("NewClient(WithCookieJar(nil)) returned nil error, want error")
	}
}

func TestWithUserAgent_Success(t *testing.T) {
	t.Parallel()
	r := testRequest(t, mangoplus.WithUserAgent("MyAgent/9.9"))
	testHeader(t, r, "User-Agent", "MyAgent/9.9")
}

func TestWithOSVersion_Success(t *testing.T) {
	t.Parallel()
	r := testRequest(t, mangoplus.WithOSVersion("34"))
	testFormValues(t, r, values{"os": "android", "os_ver": "34", "app_ver": mangoplus.DefaultAppVersion})
}

func TestWithAppVersion_Success(t *testing.T) {
	t.Parallel()
	r := testRequest(t, mangoplus.WithAppVersion("777"))
	testFormValues(t, r, values{"os": "android", "os_ver": mangoplus.DefaultOSVersion, "app_ver": "777"})
}

func TestWithSecret_Success(t *testing.T) {
	t.Parallel()
	r := testRequest(t, mangoplus.WithSecret("sekret"))
	testFormValues(t, r, values{"secret": "sekret", "os": "android", "os_ver": mangoplus.DefaultOSVersion, "app_ver": mangoplus.DefaultAppVersion})
}

// WithAndroidID is covered by TestClient_Register, which checks the device
// token derived from it.

func TestWithAndroidID_WithSecret(t *testing.T) {
	t.Parallel()
	_, err := mangoplus.NewClient(mangoplus.WithAndroidID("abcdef0123456789"), mangoplus.WithSecret("sekret"))
	if err == nil {
		t.Error("NewClient with both AndroidID and Secret returned nil error, want error")
	}
}

func TestWithAutoRegister_Success(t *testing.T) {
	t.Parallel()
	_, mux, serverURL := setup(t)

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_RegisterationData{
				RegisterationData: &proto.RegistrationData{DeviceSecret: "auto-secret"},
			},
		})
	})

	c, err := mangoplus.NewClient(
		mangoplus.WithTransport(rewriteTransport(t, serverURL)),
		mangoplus.WithAutoRegister(t.Context()),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if got, want := c.Secret(), "auto-secret"; got != want {
		t.Errorf("Secret() is %v, want %v", got, want)
	}
}

func TestWithAutoRegister_WithSecret(t *testing.T) {
	t.Parallel()
	_, mux, serverURL := setup(t)

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		t.Error("register endpoint called despite secret being provided")
	})

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("existing"),
		mangoplus.WithTransport(rewriteTransport(t, serverURL)),
		mangoplus.WithAutoRegister(t.Context()),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if got, want := c.Secret(), "existing"; got != want {
		t.Errorf("Secret() is %v, want %v", got, want)
	}
}

func TestWithAutoRegister_NilContext(t *testing.T) {
	t.Parallel()
	var nilCtx context.Context
	if _, err := mangoplus.NewClient(mangoplus.WithAutoRegister(nilCtx)); err == nil {
		t.Error("NewClient(WithAutoRegister(nil)) returned nil error, want error")
	}
}
