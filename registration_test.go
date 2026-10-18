package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestRegistration_Register_Success(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_RegisterationData{
						RegisterationData: &proto.RegistrationData{DeviceSecret: "secret-123"},
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

	got, err := c.Registration.Register(context.Background(), "token", "key")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if got.DeviceSecret != "secret-123" {
		t.Errorf("DeviceSecret = %q, want secret-123", got.DeviceSecret)
	}

	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.Method != http.MethodPut {
		t.Errorf("Method = %q, want PUT", gotReq.Method)
	}
	if gotReq.URL.Path != "/api/register" {
		t.Errorf("Path = %q, want /api/register", gotReq.URL.Path)
	}
	q := gotReq.URL.Query()
	if q.Get("device_token") != "token" {
		t.Errorf("device_token = %q, want token", q.Get("device_token"))
	}
	if q.Get("security_key") != "key" {
		t.Errorf("security_key = %q, want key", q.Get("security_key"))
	}
	if q.Has("secret") {
		t.Error("secret query must be omitted during registration")
	}
}

func TestRegistration_Register_EmptySecretReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_RegisterationData{
						RegisterationData: &proto.RegistrationData{DeviceSecret: ""},
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

	if _, err := c.Registration.Register(context.Background(), "t", "k"); err == nil {
		t.Error("Register returned nil error for empty secret, want non-nil")
	}
}

func TestRegistration_Register_ProtoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, errorResponse(&proto.ErrorResult{
			Action: proto.ErrorResult_DEFAULT,
			EnglishPopup: &proto.Popup_OSDefault{
				Subject: "Not Found",
				Body:    "nope",
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

	_, err = c.Registration.Register(context.Background(), "t", "k")
	if err == nil {
		t.Fatal("Register returned nil error, want *ProtoError")
	}
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true (err=%v)", err)
	}
}

func TestClient_Register_UsesDeviceToken(t *testing.T) {
	// Register derives device_token and security_key from the android ID.
	// Different android IDs must yield different tokens.
	received := map[string]url.Values{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received[r.URL.Query().Get("device_token")] = r.URL.Query()
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_RegisterationData{
						RegisterationData: &proto.RegistrationData{DeviceSecret: "ok"},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	for _, id := range []string{"aaaaaaaa", "bbbbbbbb"} {
		c, err := mangoplus.NewClient(
			mangoplus.WithAndroidID(id),
			mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
		)
		if err != nil {
			t.Fatalf("NewClient(%q): %v", id, err)
		}
		if _, err := c.Register(context.Background()); err != nil {
			t.Fatalf("Register(%q): %v", id, err)
		}
	}

	if len(received) != 2 {
		t.Errorf("got %d distinct device_tokens, want 2 (received=%v)", len(received), received)
	}
}
