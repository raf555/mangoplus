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

func TestTitle_GetTitleDetailV3_Success(t *testing.T) {
	const (
		startTS = int64(1_700_000_000)
		endTS   = int64(1_700_100_000)
		nextTS  = int64(1_700_200_000)
	)

	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		writeProtoResponse(t, w, &proto.Response{
			Result: &proto.Response_Success{
				Success: &proto.SuccessResult{
					Data: &proto.SuccessResult_TitleDetailView{
						TitleDetailView: &proto.TitleDetailView{
							Title: &proto.Title{
								TitleId:  42,
								Name:     "The Title",
								Author:   "Someone",
								Language: proto.Language_INDONESIAN,
							},
							TitleImageUrl:      "https://img/title",
							Overview:           "overview text",
							BackgroundImageUrl: "https://img/bg",
							NextTimeStamp:      uint32(nextTS),
							IsSimulReleased:    true,
							Rating:             proto.TitleDetailView_TEENPLUS,
							NumberOfViews:      123,
							TitleLanguages: []*proto.TitleDetailView_TitleLanguages{
								{TitleId: 42, Language: proto.Language_ENGLISH},
								{TitleId: 100, Language: proto.Language_INDONESIAN},
							},
							Tags: []*proto.Tag{
								{Tag: "Action", Slug: "action"},
								{Tag: "Comedy", Slug: "comedy"},
							},
							TitleLabels: &proto.TitleDetailView_TitleLabels{
								ReleaseSchedule: proto.TitleDetailView_WEEKLY,
								IsSimulpub:      true,
							},
							Label: &proto.Label{
								Label:       proto.LabelCodes_WJ,
								Description: "Weekly Jump",
							},
							ChapterListGroup: []*proto.TitleDetailView_ChapterGroup{
								{
									ChapterNumbers: "1-10",
									FirstChapterList: []*proto.Chapter{
										{
											TitleId:        42,
											ChapterId:      1,
											Name:           "Chapter 1",
											SubTitle:       "Beginnings",
											ThumbnailUrl:   "https://img/thumb",
											StartTimeStamp: uint32(startTS),
											EndTimeStamp:   uint32(endTS),
											ChapterType:    proto.ChapterType_FREE,
											ViewCount:      1000,
										},
									},
									MidChapterList: []*proto.Chapter{
										{ChapterId: 5, ChapterType: proto.ChapterType_STANDARD},
									},
									LastChapterList: []*proto.Chapter{
										{ChapterId: 10, ChapterType: proto.ChapterType_DELUXE},
									},
								},
							},
							ChapterListV2: []*proto.Chapter{
								{ChapterId: 1, ChapterType: proto.ChapterType_FREE},
								{ChapterId: 10, ChapterType: proto.ChapterType_LOCKED_AFTER_FREE_READ},
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

	view, err := c.Title.GetTitleDetailV3(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetTitleDetailV3: %v", err)
	}

	// Request shape.
	if gotReq.URL.Path != "/api/title_detailV3" {
		t.Errorf("Path = %q, want /api/title_detailV3", gotReq.URL.Path)
	}
	if got := gotReq.URL.Query().Get("title_id"); got != strconv.Itoa(42) {
		t.Errorf("title_id query = %q, want 42", got)
	}

	// Top-level fields.
	if view.Title.TitleID != 42 {
		t.Errorf("Title.TitleID = %d, want 42", view.Title.TitleID)
	}
	if view.Title.Language != mangoplus.LanguageIndonesian {
		t.Errorf("Title.Language = %q, want %q", view.Title.Language, mangoplus.LanguageIndonesian)
	}
	if view.ImageURL != "https://img/title" {
		t.Errorf("ImageURL = %q, want https://img/title", view.ImageURL)
	}
	if view.TitleImageURL != "https://img/title" {
		t.Errorf("TitleImageURL = %q, want https://img/title", view.TitleImageURL)
	}
	if view.BackgroundImageURL != "https://img/bg" {
		t.Errorf("BackgroundImageURL = %q, want https://img/bg", view.BackgroundImageURL)
	}
	if !view.NextTimestamp.Equal(time.Unix(nextTS, 0)) {
		t.Errorf("NextTimestamp = %v, want %v", view.NextTimestamp, time.Unix(nextTS, 0))
	}
	if !view.IsSimulRelease {
		t.Error("IsSimulRelease = false, want true")
	}
	if view.Rating != mangoplus.RatingTeenPlus {
		t.Errorf("Rating = %q, want %q", view.Rating, mangoplus.RatingTeenPlus)
	}
	if view.NumberOfViews != 123 {
		t.Errorf("NumberOfViews = %d, want 123", view.NumberOfViews)
	}

	// Languages.
	if len(view.Languages) != 2 {
		t.Fatalf("len(Languages) = %d, want 2", len(view.Languages))
	}
	if view.Languages[0].TitleID != 42 || view.Languages[0].Language != mangoplus.LanguageEnglish {
		t.Errorf("Languages[0] = %+v, want {TitleID:42 Language:ENGLISH}", view.Languages[0])
	}

	// Tags.
	wantTags := []mangoplus.Tag{
		{Name: "Action", Slug: "action"},
		{Name: "Comedy", Slug: "comedy"},
	}
	if len(view.Tags) != len(wantTags) {
		t.Fatalf("len(Tags) = %d, want %d", len(view.Tags), len(wantTags))
	}
	for i, wt := range wantTags {
		if view.Tags[i] != wt {
			t.Errorf("Tags[%d] = %+v, want %+v", i, view.Tags[i], wt)
		}
	}

	// Labels.
	if view.Labels.ReleaseSchedule != mangoplus.ReleaseScheduleWeekly {
		t.Errorf("Labels.ReleaseSchedule = %q, want %q", view.Labels.ReleaseSchedule, mangoplus.ReleaseScheduleWeekly)
	}
	if !view.Labels.IsSimulpub {
		t.Error("Labels.IsSimulpub = false, want true")
	}
	if view.Label.Code != mangoplus.LabelCodeWJ {
		t.Errorf("Label.Code = %q, want %q", view.Label.Code, mangoplus.LabelCodeWJ)
	}
	if view.Label.Description != "Weekly Jump" {
		t.Errorf("Label.Description = %q, want Weekly Jump", view.Label.Description)
	}

	// Chapter list group.
	if len(view.ChapterListGroup) != 1 {
		t.Fatalf("len(ChapterListGroup) = %d, want 1", len(view.ChapterListGroup))
	}
	grp := view.ChapterListGroup[0]
	if grp.ChapterNumbers != "1-10" {
		t.Errorf("ChapterNumbers = %q, want 1-10", grp.ChapterNumbers)
	}
	if len(grp.FirstChapterList) != 1 {
		t.Fatalf("len(FirstChapterList) = %d, want 1", len(grp.FirstChapterList))
	}
	fc := grp.FirstChapterList[0]
	if fc.ChapterID != 1 || fc.Name != "Chapter 1" || fc.SubTitle != "Beginnings" {
		t.Errorf("FirstChapterList[0] = %+v", fc)
	}
	if !fc.StartTimestamp.Equal(time.Unix(startTS, 0)) {
		t.Errorf("StartTimestamp = %v, want %v", fc.StartTimestamp, time.Unix(startTS, 0))
	}
	if !fc.EndTimestamp.Equal(time.Unix(endTS, 0)) {
		t.Errorf("EndTimestamp = %v, want %v", fc.EndTimestamp, time.Unix(endTS, 0))
	}
	if fc.ChapterType != mangoplus.ChapterTypeFree {
		t.Errorf("ChapterType = %q, want %q", fc.ChapterType, mangoplus.ChapterTypeFree)
	}
	if fc.ViewCount != 1000 {
		t.Errorf("ViewCount = %d, want 1000", fc.ViewCount)
	}
	if grp.MidChapterList[0].ChapterType != mangoplus.ChapterTypeStandard {
		t.Errorf("MidChapterList[0].ChapterType = %q, want %q",
			grp.MidChapterList[0].ChapterType, mangoplus.ChapterTypeStandard)
	}
	if grp.LastChapterList[0].ChapterType != mangoplus.ChapterTypeDeluxe {
		t.Errorf("LastChapterList[0].ChapterType = %q, want %q",
			grp.LastChapterList[0].ChapterType, mangoplus.ChapterTypeDeluxe)
	}

	// ChapterListV2 flat list.
	if len(view.ChapterListV2) != 2 {
		t.Fatalf("len(ChapterListV2) = %d, want 2", len(view.ChapterListV2))
	}
	if view.ChapterListV2[1].ChapterType != mangoplus.ChapterTypeLockedAfterFreeRead {
		t.Errorf("ChapterListV2[1].ChapterType = %q, want %q",
			view.ChapterListV2[1].ChapterType, mangoplus.ChapterTypeLockedAfterFreeRead)
	}
}

func TestTitle_GetTitleDetailV3_RatingValues(t *testing.T) {
	tests := []struct {
		name string
		pb   proto.TitleDetailView_Rating
		want mangoplus.Rating
	}{
		{"all ages", proto.TitleDetailView_ALLAGE, mangoplus.RatingAllAges},
		{"teen", proto.TitleDetailView_TEEN, mangoplus.RatingTeen},
		{"teen plus", proto.TitleDetailView_TEENPLUS, mangoplus.RatingTeenPlus},
		{"mature", proto.TitleDetailView_MATURE, mangoplus.RatingMature},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProtoResponse(t, w, &proto.Response{
					Result: &proto.Response_Success{
						Success: &proto.SuccessResult{
							Data: &proto.SuccessResult_TitleDetailView{
								TitleDetailView: &proto.TitleDetailView{Rating: tt.pb},
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

			view, err := c.Title.GetTitleDetailV3(context.Background(), 1)
			if err != nil {
				t.Fatalf("GetTitleDetailV3: %v", err)
			}
			if view.Rating != tt.want {
				t.Errorf("Rating = %q, want %q", view.Rating, tt.want)
			}
		})
	}
}

func TestTitle_GetTitleDetailV3_ReleaseScheduleValues(t *testing.T) {
	tests := []struct {
		pb   proto.TitleDetailView_ReleaseSchedule
		want mangoplus.ReleaseSchedule
	}{
		{proto.TitleDetailView_DISABLED, mangoplus.ReleaseScheduleDisabled},
		{proto.TitleDetailView_EVERYDAY, mangoplus.ReleaseScheduleEveryday},
		{proto.TitleDetailView_WEEKLY, mangoplus.ReleaseScheduleWeekly},
		{proto.TitleDetailView_BIWEEKLY, mangoplus.ReleaseScheduleBiweekly},
		{proto.TitleDetailView_MONTHLY, mangoplus.ReleaseScheduleMonthly},
		{proto.TitleDetailView_BIMONTHLY, mangoplus.ReleaseScheduleBiMonthly},
		{proto.TitleDetailView_TRIMONTHLY, mangoplus.ReleaseScheduleTriMonthly},
		{proto.TitleDetailView_OTHER, mangoplus.ReleaseScheduleOther},
		{proto.TitleDetailView_COMPLETED, mangoplus.ReleaseScheduleCompleted},
		{proto.TitleDetailView_ONE_SHOT, mangoplus.ReleaseScheduleOneShot},
		{proto.TitleDetailView_HIATUS, mangoplus.ReleaseScheduleHiatus},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProtoResponse(t, w, &proto.Response{
					Result: &proto.Response_Success{
						Success: &proto.SuccessResult{
							Data: &proto.SuccessResult_TitleDetailView{
								TitleDetailView: &proto.TitleDetailView{
									TitleLabels: &proto.TitleDetailView_TitleLabels{ReleaseSchedule: tt.pb},
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

			view, err := c.Title.GetTitleDetailV3(context.Background(), 1)
			if err != nil {
				t.Fatalf("GetTitleDetailV3: %v", err)
			}
			if view.Labels.ReleaseSchedule != tt.want {
				t.Errorf("ReleaseSchedule = %q, want %q", view.Labels.ReleaseSchedule, tt.want)
			}
		})
	}
}

func TestTitle_GetTitleDetailV3_LabelCodes(t *testing.T) {
	tests := []struct {
		pb   proto.LabelCodes
		want mangoplus.LabelCode
	}{
		{proto.LabelCodes_WJ, mangoplus.LabelCodeWJ},
		{proto.LabelCodes_SQ, mangoplus.LabelCodeSQ},
		{proto.LabelCodes_VJ, mangoplus.LabelCodeVJ},
		{proto.LabelCodes_YJ, mangoplus.LabelCodeYJ},
		{proto.LabelCodes_J_PLUS, mangoplus.LabelCodeJPlus},
		{proto.LabelCodes_REVIVAL, mangoplus.LabelCodeRevival},
		{proto.LabelCodes_CREATORS, mangoplus.LabelCodeCreators},
		{proto.LabelCodes_MEE, mangoplus.LabelCodeMEE},
		{proto.LabelCodes_TYJ, mangoplus.LabelCodeTYJ},
		{proto.LabelCodes_OTHERS, mangoplus.LabelCodeOthers},
		{proto.LabelCodes_SKJ, mangoplus.LabelCodeSKJ},
		{proto.LabelCodes_GIGA, mangoplus.LabelCodeGiga},
		{proto.LabelCodes_UJ, mangoplus.LabelCodeUJ},
		{proto.LabelCodes_DX, mangoplus.LabelCodeDX},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeProtoResponse(t, w, &proto.Response{
					Result: &proto.Response_Success{
						Success: &proto.SuccessResult{
							Data: &proto.SuccessResult_TitleDetailView{
								TitleDetailView: &proto.TitleDetailView{
									Label: &proto.Label{Label: tt.pb, Description: "d"},
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

			view, err := c.Title.GetTitleDetailV3(context.Background(), 1)
			if err != nil {
				t.Fatalf("GetTitleDetailV3: %v", err)
			}
			if view.Label.Code != tt.want {
				t.Errorf("Label.Code = %q, want %q", view.Label.Code, tt.want)
			}
		})
	}
}

func TestTitle_GetTitleDetailV3_NilViewReturnsError(t *testing.T) {
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

	if _, err := c.Title.GetTitleDetailV3(context.Background(), 1); err == nil {
		t.Error("GetTitleDetailV3 returned nil error for empty success body, want non-nil")
	}
}

func TestTitle_GetTitleDetailV3_ProtoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeProtoResponse(t, w, errorResponse(&proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Invalid Parameter", Body: "bad id"},
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

	_, err = c.Title.GetTitleDetailV3(context.Background(), 0)
	if err == nil {
		t.Fatal("GetTitleDetailV3 returned nil error, want *ProtoError")
	}
	if !errors.Is(err, mangoplus.ErrInvalidParameter) {
		t.Errorf("errors.Is(err, ErrInvalidParameter) = false, want true (err=%v)", err)
	}
}
