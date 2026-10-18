package mangoplus_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raf555/mangoplus"
	"github.com/raf555/mangoplus/internal/proto"
)

func TestTitleList_AllV2_Success(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_AllTitlesViewV2{
						AllTitlesViewV2: &proto.AllTitlesViewV2{
							AllTitlesGroup: []*proto.AllTitlesGroup{
								{
									TheTitle: "A",
									Titles: []*proto.Title{
										{
											TitleId:          1,
											Name:             "Alpha",
											Author:           "Author A",
											PortraitImageUrl: "https://img/portrait",
											LandscapeImageUrl: "https://img/landscape",
											Language:         proto.Language_ENGLISH,
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

	view, err := c.TitleList.AllV2(context.Background())
	if err != nil {
		t.Fatalf("AllV2: %v", err)
	}

	if gotReq == nil {
		t.Fatal("no request captured")
	}
	if gotReq.Method != http.MethodGet {
		t.Errorf("Method = %q, want GET", gotReq.Method)
	}
	if gotReq.URL.Path != "/api/title_list/allV2" {
		t.Errorf("Path = %q, want /api/title_list/allV2", gotReq.URL.Path)
	}

	if len(view.AllTitlesGroup) != 2 {
		t.Fatalf("len(AllTitlesGroup) = %d, want 2", len(view.AllTitlesGroup))
	}

	g0 := view.AllTitlesGroup[0]
	if g0.TheTitle != "A" {
		t.Errorf("Group[0].TheTitle = %q, want A", g0.TheTitle)
	}
	if len(g0.Titles) != 1 {
		t.Fatalf("len(Group[0].Titles) = %d, want 1", len(g0.Titles))
	}
	title := g0.Titles[0]
	if title.TitleID != 1 {
		t.Errorf("TitleID = %d, want 1", title.TitleID)
	}
	if title.Name != "Alpha" {
		t.Errorf("Name = %q, want Alpha", title.Name)
	}
	if title.Author != "Author A" {
		t.Errorf("Author = %q, want Author A", title.Author)
	}
	if title.PortraitImageURL != "https://img/portrait" {
		t.Errorf("PortraitImageURL = %q, want https://img/portrait", title.PortraitImageURL)
	}
	if title.LandscapeImageURL != "https://img/landscape" {
		t.Errorf("LandscapeImageURL = %q, want https://img/landscape", title.LandscapeImageURL)
	}
	if title.Language != mangoplus.LanguageEnglish {
		t.Errorf("Language = %q, want %q", title.Language, mangoplus.LanguageEnglish)
	}

	g1 := view.AllTitlesGroup[1]
	if len(g1.Titles) != 2 {
		t.Fatalf("len(Group[1].Titles) = %d, want 2", len(g1.Titles))
	}
	if g1.Titles[0].Language != mangoplus.LanguageFrench {
		t.Errorf("Group[1].Titles[0].Language = %q, want %q", g1.Titles[0].Language, mangoplus.LanguageFrench)
	}
	if g1.Titles[1].Language != mangoplus.LanguageSpanish {
		t.Errorf("Group[1].Titles[1].Language = %q, want %q", g1.Titles[1].Language, mangoplus.LanguageSpanish)
	}
}

func TestTitleList_AllV2_NilViewReturnsError(t *testing.T) {
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

	if _, err := c.TitleList.AllV2(context.Background()); err == nil {
		t.Error("AllV2 returned nil error for empty success body, want non-nil")
	}
}

func TestTitleList_AllV2_PropagatesProtoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, errorResponse(&proto.ErrorResult{
			Action: proto.ErrorResult_MAINTENANCE,
			EnglishPopup: &proto.Popup_OSDefault{
				Subject: "NEW VERSION AVAILABLE",
				Body:    "update app",
			},
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

	_, err = c.TitleList.AllV2(context.Background())
	if err == nil {
		t.Fatal("AllV2 returned nil error, want *ProtoError")
	}
	if !errors.Is(err, mangoplus.ErrNewVersionAvailable) {
		t.Errorf("errors.Is(err, ErrNewVersionAvailable) = false, want true (err=%v)", err)
	}

	var pErr *mangoplus.ProtoError
	if !errors.As(err, &pErr) {
		t.Fatalf("err type = %T, want *mangoplus.ProtoError", err)
	}
	if pErr.Action != mangoplus.ErrorActionMaintenance {
		t.Errorf("Action = %q, want %q", pErr.Action, mangoplus.ErrorActionMaintenance)
	}
}

func TestTitleList_AllV2_RequestCarriesClientConfig(t *testing.T) {
	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_AllTitlesViewV2{
						AllTitlesViewV2: &proto.AllTitlesViewV2{},
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := mangoplus.NewClient(
		mangoplus.WithSecret("my-secret"),
		mangoplus.WithOSVersion("35"),
		mangoplus.WithAppVersion("300"),
		mangoplus.WithTransport(rewriteTransport(t, srv.URL)),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := c.TitleList.AllV2(context.Background()); err != nil {
		t.Fatalf("AllV2: %v", err)
	}

	q := gotReq.URL.Query()
	if q.Get("secret") != "my-secret" {
		t.Errorf("secret = %q, want my-secret", q.Get("secret"))
	}
	if q.Get("os") != "android" {
		t.Errorf("os = %q, want android", q.Get("os"))
	}
	if q.Get("os_ver") != "35" {
		t.Errorf("os_ver = %q, want 35", q.Get("os_ver"))
	}
	if q.Get("app_ver") != "300" {
		t.Errorf("app_ver = %q, want 300", q.Get("app_ver"))
	}
}
