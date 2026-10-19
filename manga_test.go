package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

// testViewChapter serves viewer from the manga viewer endpoint and returns
// what Manga.ViewChapter makes of it.
func testViewChapter(t *testing.T, viewer *proto.MangaViewer) mangoplus.MangaViewer {
	t.Helper()
	client, mux, _ := setup(t)

	mux.HandleFunc("/manga_viewer", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_MangaViewer{MangaViewer: viewer},
		})
	})

	got, err := client.Manga.ViewChapter(t.Context(), 1, mangoplus.DefaultViewChapterOptions())
	if err != nil {
		t.Fatalf("Manga.ViewChapter returned error: %v", err)
	}
	return got
}

func TestDefaultViewChapterOptions(t *testing.T) {
	t.Parallel()
	got := mangoplus.DefaultViewChapterOptions()
	want := mangoplus.ViewChapterOptions{SplitImages: false, ImageQuality: mangoplus.ImageQualitySuperHigh}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DefaultViewChapterOptions mismatch (-want +got):\n%s", diff)
	}
}

func TestMangaService_ViewChapter_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	const (
		startTS = 1_700_000_000
		endTS   = 1_700_100_000
	)

	mux.HandleFunc("/manga_viewer", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"chapter_id":  "99",
			"split":       "yes",
			"img_quality": "high",
			"secret":      "s",
			"os":          "android",
			"os_ver":      mangoplus.DefaultOSVersion,
			"app_ver":     mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_MangaViewer{
				MangaViewer: &proto.MangaViewer{
					ChapterId:        99,
					TitleId:          42,
					TitleName:        "Alpha",
					ChapterName:      "Chapter 99",
					NumberOfComments: 7,
					StartFromRight:   true,
					IsVerticalOnly:   false,
					IsHorizontalOnly: true,
					TitleLanguage:    "eng",
					TitleAvailableLanguages: []*proto.MangaViewer_TitleAvailableLanguages{
						{TitleId: 42, Language: proto.Language_ENGLISH},
						{TitleId: 55, Language: proto.Language_INDONESIAN},
					},
					Pages: []*proto.Page{
						{
							Data: &proto.Page_MangaPage_{
								MangaPage: &proto.Page_MangaPage{
									ImageUrl:      "https://img/p1",
									Width:         720,
									Height:        1024,
									Type:          proto.Page_SINGLE,
									EncryptionKey: "key1",
								},
							},
						},
						{
							Data: &proto.Page_MangaPage_{
								MangaPage: &proto.Page_MangaPage{
									ImageUrl:      "https://img/p2",
									Width:         1440,
									Height:        1024,
									Type:          proto.Page_DOUBLE,
									EncryptionKey: "key2",
								},
							},
						},
						// Only MangaPages are kept; the LastPage provides CurrentChapter.
						{
							Data: &proto.Page_LastPage_{
								LastPage: &proto.Page_LastPage{
									CurrentChapter: &proto.Chapter{
										TitleId:        42,
										ChapterId:      99,
										Name:           "Chapter 99",
										StartTimeStamp: startTS,
										EndTimeStamp:   endTS,
										ChapterType:    proto.ChapterType_FREE,
									},
								},
							},
						},
					},
				},
			},
		})
	})

	opts := mangoplus.ViewChapterOptions{SplitImages: true, ImageQuality: mangoplus.ImageQualityHigh}
	got, err := client.Manga.ViewChapter(t.Context(), 99, opts)
	if err != nil {
		t.Fatalf("Manga.ViewChapter returned error: %v", err)
	}

	want := mangoplus.MangaViewer{
		MangaPages: []mangoplus.MangaPage{
			{ImageURL: "https://img/p1", Width: 720, Height: 1024, Type: mangoplus.PageTypeSingle, EncryptionKey: "key1"},
			{ImageURL: "https://img/p2", Width: 1440, Height: 1024, Type: mangoplus.PageTypeDouble, EncryptionKey: "key2"},
		},
		CurrentChapter: mangoplus.Chapter{
			TitleID:        42,
			ChapterID:      99,
			Name:           "Chapter 99",
			StartTimestamp: time.Unix(startTS, 0),
			EndTimestamp:   time.Unix(endTS, 0),
			ChapterType:    mangoplus.ChapterTypeFree,
		},
		ChapterID:        99,
		TitleName:        "Alpha",
		ChapterName:      "Chapter 99",
		NumberOfComments: 7,
		TitleID:          42,
		TitleLanguage:    mangoplus.LanguageEnglish,
		AvailableLanguages: []mangoplus.MangaAvailableLanguage{
			{TitleID: 42, Language: mangoplus.LanguageEnglish},
			{TitleID: 55, Language: mangoplus.LanguageIndonesian},
		},
		StartFromRight:   true,
		IsVerticalOnly:   false,
		IsHorizontalOnly: true,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Manga.ViewChapter mismatch (-want +got):\n%s", diff)
	}
}

func TestMangaService_ViewChapter_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Manga.ViewChapter", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Manga.ViewChapter(ctx, 99, mangoplus.DefaultViewChapterOptions())
		return err
	})
}

func TestMangaService_ViewChapter_ZeroOptions(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/manga_viewer", func(w http.ResponseWriter, r *http.Request) {
		// Zero options mean no split and super_high quality.
		testFormValues(t, r, values{
			"chapter_id":  "1",
			"split":       "no",
			"img_quality": "super_high",
			"os":          "android",
			"os_ver":      mangoplus.DefaultOSVersion,
			"app_ver":     mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_MangaViewer{MangaViewer: &proto.MangaViewer{}},
		})
	})

	if _, err := client.Manga.ViewChapter(t.Context(), 1, mangoplus.ViewChapterOptions{}); err != nil {
		t.Errorf("Manga.ViewChapter returned error: %v", err)
	}
}

func TestMangaService_ViewChapter_NilViewer(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/manga_viewer", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Manga.ViewChapter(t.Context(), 1, mangoplus.DefaultViewChapterOptions()); err == nil {
		t.Error("Manga.ViewChapter returned nil error for empty success body, want error")
	}
}

func TestMangaService_ViewChapter_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/manga_viewer", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Not Found", Body: "nope"},
		})
	})

	_, err := client.Manga.ViewChapter(t.Context(), 1, mangoplus.DefaultViewChapterOptions())
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("Manga.ViewChapter error = %v, want ErrNotFound", err)
	}
}

func TestMangaService_ViewChapter_PageType(t *testing.T) {
	t.Parallel()
	page := func(pt proto.Page_PageType) *proto.Page {
		return &proto.Page{Data: &proto.Page_MangaPage_{MangaPage: &proto.Page_MangaPage{Type: pt}}}
	}
	got := testViewChapter(t, &proto.MangaViewer{
		Pages: []*proto.Page{
			page(proto.Page_SINGLE),
			page(proto.Page_LEFT),
			page(proto.Page_RIGHT),
			page(proto.Page_DOUBLE),
			page(proto.Page_PageType(-1)),
		},
	})

	want := []mangoplus.PageType{
		mangoplus.PageTypeSingle,
		mangoplus.PageTypeLeft,
		mangoplus.PageTypeRight,
		mangoplus.PageTypeDouble,
		"",
	}
	var gotTypes []mangoplus.PageType
	for _, p := range got.MangaPages {
		gotTypes = append(gotTypes, p.Type)
	}
	if diff := cmp.Diff(want, gotTypes); diff != "" {
		t.Errorf("MangaPages page types mismatch (-want +got):\n%s", diff)
	}
}

func TestMangaService_ViewChapter_TitleLanguage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lang string
		want mangoplus.Language
	}{
		{"eng", mangoplus.LanguageEnglish},
		{"esp", mangoplus.LanguageSpanish},
		{"fra", mangoplus.LanguageFrench},
		{"ind", mangoplus.LanguageIndonesian},
		{"ptb", mangoplus.LanguagePortugueseBR},
		{"rus", mangoplus.LanguageRussian},
		{"tha", mangoplus.LanguageThai},
		{"vie", mangoplus.LanguageVietnamese},
		{"deu", mangoplus.LanguageGerman},
		{"xxx", ""},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			t.Parallel()
			got := testViewChapter(t, &proto.MangaViewer{TitleLanguage: tt.lang})
			if got.TitleLanguage != tt.want {
				t.Errorf("TitleLanguage is %q, want %q", got.TitleLanguage, tt.want)
			}
		})
	}
}
