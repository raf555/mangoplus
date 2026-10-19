package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestRegistrationService_Register_Success(t *testing.T) {
	t.Parallel()
	// A configured secret must not be sent while registering.
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testFormValues(t, r, values{
			"device_token": "token",
			"security_key": "key",
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

	got, err := client.Registration.Register(t.Context(), "token", "key")
	if err != nil {
		t.Fatalf("Registration.Register returned error: %v", err)
	}

	want := mangoplus.RegistrationData{DeviceSecret: "secret-123"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Registration.Register mismatch (-want +got):\n%s", diff)
	}
}

func TestRegistrationService_Register_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Registration.Register", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Registration.Register(ctx, "token", "key")
		return err
	})
}

func TestRegistrationService_Register_EmptySecret(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_RegisterationData{
				RegisterationData: &proto.RegistrationData{DeviceSecret: ""},
			},
		})
	})

	if _, err := client.Registration.Register(t.Context(), "token", "key"); err == nil {
		t.Error("Registration.Register returned nil error for empty secret, want error")
	}
}

func TestRegistrationService_Register_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Not Found", Body: "nope"},
		})
	})

	_, err := client.Registration.Register(t.Context(), "token", "key")
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("Registration.Register error = %v, want ErrNotFound", err)
	}
}
