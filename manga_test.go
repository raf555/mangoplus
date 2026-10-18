package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestDefaultViewChapterOptions(t *testing.T) {
	got := mangoplus.DefaultViewChapterOptions()
	if got.SplitImages {
		t.Error("default SplitImages = true, want false")
	}
	if got.ImageQuality != mangoplus.ImageQualitySuperHigh {
		t.Errorf("default ImageQuality = %q, want %q", got.ImageQuality, mangoplus.ImageQualitySuperHigh)
	}
}

func TestManga_ViewChapter_Success(t *testing.T) {
	const (
		startTS = int64(1_700_000_000)
		endTS   = int64(1_700_100_000)
	)

	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
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
								// A LastPage with CurrentChapter to exercise
								// that branch.
								{
									Data: &proto.Page_LastPage_{
										LastPage: &proto.Page_LastPage{
											CurrentChapter: &proto.Chapter{
												TitleId:        42,
												ChapterId:      99,
												Name:           "Chapter 99",
												StartTimeStamp: uint32(startTS),
												EndTimeStamp:   uint32(endTS),
												ChapterType:    proto.ChapterType_FREE,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	view, err := c.Manga.ViewChapter(context.Background(), 99, mangoplus.ViewChapterOptions{
		SplitImages:  true,
		ImageQuality: mangoplus.ImageQualityHigh,
	})
	if err != nil {
		t.Fatalf("ViewChapter: %v", err)
	}

	// Request shape.
	if gotReq.URL.Path != "/api/manga_viewer" {
		t.Errorf("Path = %q, want /api/manga_viewer", gotReq.URL.Path)
	}
	q := gotReq.URL.Query()
	if q.Get("chapter_id") != strconv.Itoa(99) {
		t.Errorf("chapter_id = %q, want 99", q.Get("chapter_id"))
	}
	if q.Get("split") != "yes" {
		t.Errorf("split = %q, want yes", q.Get("split"))
	}
	if q.Get("img_quality") != string(mangoplus.ImageQualityHigh) {
		t.Errorf("img_quality = %q, want %q", q.Get("img_quality"), mangoplus.ImageQualityHigh)
	}

	// Response shape.
	if view.ChapterID != 99 || view.TitleID != 42 {
		t.Errorf("ChapterID/TitleID = %d/%d, want 99/42", view.ChapterID, view.TitleID)
	}
	if view.TitleName != "Alpha" || view.ChapterName != "Chapter 99" {
		t.Errorf("TitleName/ChapterName = %q/%q", view.TitleName, view.ChapterName)
	}
	if view.NumberOfComments != 7 {
		t.Errorf("NumberOfComments = %d, want 7", view.NumberOfComments)
	}
	if !view.StartFromRight {
		t.Error("StartFromRight = false, want true")
	}
	if !view.IsHorizontalOnly {
		t.Error("IsHorizontalOnly = false, want true")
	}
	if view.IsVerticalOnly {
		t.Error("IsVerticalOnly = true, want false")
	}
	if view.TitleLanguage != mangoplus.LanguageEnglish {
		t.Errorf("TitleLanguage = %q, want %q", view.TitleLanguage, mangoplus.LanguageEnglish)
	}

	if len(view.AvailableLanguages) != 2 {
		t.Fatalf("len(AvailableLanguages) = %d, want 2", len(view.AvailableLanguages))
	}
	if view.AvailableLanguages[1].TitleID != 55 || view.AvailableLanguages[1].Language != mangoplus.LanguageIndonesian {
		t.Errorf("AvailableLanguages[1] = %+v", view.AvailableLanguages[1])
	}

	// Pages: only MangaPages are retained, LastPages are skipped.
	if len(view.MangaPages) != 2 {
		t.Fatalf("len(MangaPages) = %d, want 2", len(view.MangaPages))
	}
	p0 := view.MangaPages[0]
	wantP0 := mangoplus.MangaPage{
		ImageURL:      "https://img/p1",
		Width:         720,
		Height:        1024,
		Type:          mangoplus.PageTypeSingle,
		EncryptionKey: "key1",
	}
	if p0 != wantP0 {
		t.Errorf("MangaPages[0] = %+v, want %+v", p0, wantP0)
	}
	if view.MangaPages[1].Type != mangoplus.PageTypeDouble {
		t.Errorf("MangaPages[1].Type = %q, want %q", view.MangaPages[1].Type, mangoplus.PageTypeDouble)
	}

	// CurrentChapter is populated from LastPage.
	if view.CurrentChapter.ChapterID != 99 {
		t.Errorf("CurrentChapter.ChapterID = %d, want 99", view.CurrentChapter.ChapterID)
	}
	if view.CurrentChapter.ChapterType != mangoplus.ChapterTypeFree {
		t.Errorf("CurrentChapter.ChapterType = %q, want %q", view.CurrentChapter.ChapterType, mangoplus.ChapterTypeFree)
	}
	if !view.CurrentChapter.StartTimestamp.Equal(time.Unix(startTS, 0)) {
		t.Errorf("CurrentChapter.StartTimestamp = %v, want %v", view.CurrentChapter.StartTimestamp, time.Unix(startTS, 0))
	}
}

func TestManga_ViewChapter_SplitNo(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_MangaViewer{
						MangaViewer: &proto.MangaViewer{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.Manga.ViewChapter(context.Background(), 1, mangoplus.ViewChapterOptions{}); err != nil {
		t.Fatalf("ViewChapter: %v", err)
	}
	if got := gotReq.URL.Query().Get("split"); got != "no" {
		t.Errorf("split = %q, want no", got)
	}
}

func TestManga_ViewChapter_DefaultImageQuality(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_MangaViewer{
						MangaViewer: &proto.MangaViewer{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// Empty ImageQuality string should default to super_high.
	if _, err := c.Manga.ViewChapter(context.Background(), 1, mangoplus.ViewChapterOptions{}); err != nil {
		t.Fatalf("ViewChapter: %v", err)
	}
	if got := gotReq.URL.Query().Get("img_quality"); got != string(mangoplus.ImageQualitySuperHigh) {
		t.Errorf("img_quality = %q, want %q", got, mangoplus.ImageQualitySuperHigh)
	}
}

func TestManga_ViewChapter_PageTypes(t *testing.T) {
	tests := []struct {
		pb   proto.Page_PageType
		want mangoplus.PageType
	}{
		{proto.Page_SINGLE, mangoplus.PageTypeSingle},
		{proto.Page_LEFT, mangoplus.PageTypeLeft},
		{proto.Page_RIGHT, mangoplus.PageTypeRight},
		{proto.Page_DOUBLE, mangoplus.PageTypeDouble},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProtoResponse(t, w, &proto.Response{
					Result: &proto.Response_Success{
						Success: &proto.SuccessResult{
							Data: &proto.SuccessResult_MangaViewer{
								MangaViewer: &proto.MangaViewer{
									Pages: []*proto.Page{
										{Data: &proto.Page_MangaPage_{MangaPage: &proto.Page_MangaPage{Type: tt.pb}}},
									},
								},
							},
						},
					},
				})
			}))
			t.Cleanup(srv.Close)

			c, err := mangoplus.NewClient(
				mangoplus.WithSecret("s"),
				mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
			)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			view, err := c.Manga.ViewChapter(context.Background(), 1, mangoplus.DefaultViewChapterOptions())
			if err != nil {
				t.Fatalf("ViewChapter: %v", err)
			}
			if len(view.MangaPages) != 1 {
				t.Fatalf("len(MangaPages) = %d, want 1", len(view.MangaPages))
			}
			if view.MangaPages[0].Type != tt.want {
				t.Errorf("Type = %q, want %q", view.MangaPages[0].Type, tt.want)
			}
		})
	}
}

func TestManga_ViewChapter_LanguageMapping(t *testing.T) {
	tests := []struct {
		titleLang string
		want      mangoplus.Language
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
		{"xxx", mangoplus.Language("")},
	}

	for _, tt := range tests {
		t.Run(tt.titleLang, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProtoResponse(t, w, &proto.Response{
					Result: &proto.Response_Success{
						Success: &proto.SuccessResult{
							Data: &proto.SuccessResult_MangaViewer{
								MangaViewer: &proto.MangaViewer{TitleLanguage: tt.titleLang},
							},
						},
					},
				})
			}))
			t.Cleanup(srv.Close)

			c, err := mangoplus.NewClient(
				mangoplus.WithSecret("s"),
				mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
			)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			view, err := c.Manga.ViewChapter(context.Background(), 1, mangoplus.DefaultViewChapterOptions())
			if err != nil {
				t.Fatalf("ViewChapter: %v", err)
			}
			if view.TitleLanguage != tt.want {
				t.Errorf("TitleLanguage = %q, want %q", view.TitleLanguage, tt.want)
			}
		})
	}
}

func TestManga_ViewChapter_NilViewerReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{Success: &proto.SuccessResult{}},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.Manga.ViewChapter(context.Background(), 1, mangoplus.DefaultViewChapterOptions()); err == nil {
		t.Error("ViewChapter returned nil error for empty success body, want non-nil")
	}
}

func TestManga_ViewChapter_ProtoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, errorResponse(&proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Not Found", Body: "nope"},
		}))
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("s"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.Manga.ViewChapter(context.Background(), 1, mangoplus.DefaultViewChapterOptions())
	if err == nil {
		t.Fatal("ViewChapter returned nil error, want *ProtoError")
	}
	if !errors.Is(err, mangoplus.ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true (err=%v)", err)
	}
}
