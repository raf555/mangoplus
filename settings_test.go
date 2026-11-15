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

func TestSettingsService_ViewV2_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		planType string
		want     mangoplus.PlanType
	}{
		{"basic", "basic", mangoplus.PlanTypeBasic},
		{"standard", "standard", mangoplus.PlanTypeStandard},
		{"deluxe", "deluxe", mangoplus.PlanTypeDeluxe},
		{"empty is basic", "", mangoplus.PlanTypeBasic},
		{"unknown is basic", "premium", mangoplus.PlanTypeBasic},
		{"match is exact", "Deluxe", mangoplus.PlanTypeBasic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, mux, _ := setup(t, mangoplus.WithSecret("s"))

			mux.HandleFunc("/settings_v2", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testFormValues(t, r, values{
					"lang":        "eng",
					"clang":       "eng",
					"viewer_mode": "horizontal",
					"secret":      "s",
					"os":          "android",
					"os_ver":      mangoplus.DefaultOSVersion,
					"app_ver":     mangoplus.DefaultAppVersion,
				})
				testWriteSuccess(t, w, &proto.SuccessResult{
					Data: &proto.SuccessResult_SettingsViewV2{
						SettingsViewV2: &proto.SettingsViewV2{
							UserName:         "reader",
							MyIcon:           &proto.CommentIcon{Id: 3, ImageUrl: "https://img/icon"},
							UserSubscription: &proto.Subscription{PlanType: tt.planType},
						},
					},
				})
			})

			got, err := client.Settings.ViewV2(t.Context())
			if err != nil {
				t.Fatalf("Settings.ViewV2 returned error: %v", err)
			}

			want := mangoplus.SettingsView{
				UserName: "reader",
				IconURL:  "https://img/icon",
				Subscription: mangoplus.Subscription{
					PlanType: tt.want,
				},
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Settings.ViewV2 mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSettingsService_ViewV2_NoSubscription(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/settings_v2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_SettingsViewV2{SettingsViewV2: &proto.SettingsViewV2{}},
		})
	})

	got, err := client.Settings.ViewV2(t.Context())
	if err != nil {
		t.Fatalf("Settings.ViewV2 returned error: %v", err)
	}
	if want := mangoplus.PlanTypeBasic; got.Subscription.PlanType != want {
		t.Errorf("Settings.ViewV2 plan type is %q, want %q", got.Subscription.PlanType, want)
	}
}

func TestSettingsService_ViewV2_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Settings.ViewV2", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Settings.ViewV2(ctx)
		return err
	})
}

func TestSettingsService_ViewV2_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/settings_v2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Settings.ViewV2(t.Context()); err == nil {
		t.Error("Settings.ViewV2 returned nil error for empty success body, want error")
	}
}

func TestSettingsService_ViewV2_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/settings_v2", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Server error", Body: "try again later"},
		})
	})

	_, err := client.Settings.ViewV2(t.Context())
	var pErr *mangoplus.ProtoError
	if !errors.As(err, &pErr) {
		t.Fatalf("Settings.ViewV2 error = %v, want *ProtoError", err)
	}
	if got, want := pErr.EnglishPopup.Subject, "Server error"; got != want {
		t.Errorf("ProtoError subject is %q, want %q", got, want)
	}
}
