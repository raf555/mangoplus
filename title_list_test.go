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

func TestTitleListService_AllV2_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"secret":  "s",
			"os":      "android",
			"os_ver":  mangoplus.DefaultOSVersion,
			"app_ver": mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_AllTitlesViewV2{
				AllTitlesViewV2: &proto.AllTitlesViewV2{
					AllTitlesGroup: []*proto.AllTitlesGroup{
						{
							TheTitle: "A",
							Titles: []*proto.Title{
								{
									TitleId:           1,
									Name:              "Alpha",
									Author:            "Author A",
									PortraitImageUrl:  "https://img/portrait",
									LandscapeImageUrl: "https://img/landscape",
									Language:          proto.Language_ENGLISH,
								},
							},
						},
						{
							TheTitle: "B",
							Titles: []*proto.Title{
								{TitleId: 2, Name: "Beta", Language: proto.Language_FRENCH},
								{TitleId: 3, Name: "Gamma", Language: proto.Language_SPANISH},
							},
						},
					},
				},
			},
		})
	})

	got, err := client.TitleList.AllV2(t.Context())
	if err != nil {
		t.Fatalf("TitleList.AllV2 returned error: %v", err)
	}

	want := mangoplus.AllTitlesViewV2{
		AllTitlesGroup: []mangoplus.AllTitlesGroupV2{
			{
				TheTitle: "A",
				Titles: []mangoplus.Title{
					{
						TitleID:           1,
						Name:              "Alpha",
						Author:            "Author A",
						PortraitImageURL:  "https://img/portrait",
						LandscapeImageURL: "https://img/landscape",
						Language:          mangoplus.LanguageEnglish,
					},
				},
			},
			{
				TheTitle: "B",
				Titles: []mangoplus.Title{
					{TitleID: 2, Name: "Beta", Language: mangoplus.LanguageFrench},
					{TitleID: 3, Name: "Gamma", Language: mangoplus.LanguageSpanish},
				},
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TitleList.AllV2 mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleListService_AllV2_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "TitleList.AllV2", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.TitleList.AllV2(ctx)
		return err
	})
}

func TestTitleListService_AllV2_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.TitleList.AllV2(t.Context()); err == nil {
		t.Error("TitleList.AllV2 returned nil error for empty success body, want error")
	}
}

func TestTitleListService_AllV2_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_list/allV2", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Not Found", Body: "nope"},
		})
	})

	_, err := client.TitleList.AllV2(t.Context())
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("TitleList.AllV2 error = %v, want ErrNotFound", err)
	}
}
