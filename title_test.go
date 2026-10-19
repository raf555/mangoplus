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

// testGetTitleDetailV3 serves view from the title detail endpoint and returns
// what Title.GetTitleDetailV3 makes of it.
func testGetTitleDetailV3(t *testing.T, view *proto.TitleDetailView) mangoplus.TitleDetailView {
	t.Helper()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_detailV3", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_TitleDetailView{TitleDetailView: view},
		})
	})

	got, err := client.Title.GetTitleDetailV3(t.Context(), 1)
	if err != nil {
		t.Fatalf("Title.GetTitleDetailV3 returned error: %v", err)
	}
	return got
}

func TestTitleService_GetTitleDetailV3_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	const (
		startTS = 1_700_000_000
		endTS   = 1_700_100_000
		nextTS  = 1_700_200_000
	)

	mux.HandleFunc("/title_detailV3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"title_id": "42",
			"secret":   "s",
			"os":       "android",
			"os_ver":   mangoplus.DefaultOSVersion,
			"app_ver":  mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
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
					NextTimeStamp:      nextTS,
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
									StartTimeStamp: startTS,
									EndTimeStamp:   endTS,
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
		})
	})

	got, err := client.Title.GetTitleDetailV3(t.Context(), 42)
	if err != nil {
		t.Fatalf("Title.GetTitleDetailV3 returned error: %v", err)
	}

	want := mangoplus.TitleDetailView{
		Title: mangoplus.Title{
			TitleID:  42,
			Name:     "The Title",
			Author:   "Someone",
			Language: mangoplus.LanguageIndonesian,
		},
		ImageURL:           "https://img/title",
		Overview:           "overview text",
		BackgroundImageURL: "https://img/bg",
		TitleImageURL:      "https://img/title",
		NextTimestamp:      time.Unix(nextTS, 0),
		IsSimulRelease:     true,
		Rating:             mangoplus.RatingTeenPlus,
		NumberOfViews:      123,
		Languages: []mangoplus.TitleLanguage{
			{TitleID: 42, Language: mangoplus.LanguageEnglish},
			{TitleID: 100, Language: mangoplus.LanguageIndonesian},
		},
		Tags: []mangoplus.Tag{
			{Name: "Action", Slug: "action"},
			{Name: "Comedy", Slug: "comedy"},
		},
		Labels: mangoplus.TitleLabels{
			ReleaseSchedule: mangoplus.ReleaseScheduleWeekly,
			IsSimulpub:      true,
		},
		Label: mangoplus.Label{
			Code:        mangoplus.LabelCodeWJ,
			Description: "Weekly Jump",
		},
		ChapterListGroup: []mangoplus.ChapterGroup{
			{
				ChapterNumbers: "1-10",
				FirstChapterList: []mangoplus.Chapter{
					{
						TitleID:        42,
						ChapterID:      1,
						Name:           "Chapter 1",
						SubTitle:       "Beginnings",
						ThumbnailURL:   "https://img/thumb",
						StartTimestamp: time.Unix(startTS, 0),
						EndTimestamp:   time.Unix(endTS, 0),
						ChapterType:    mangoplus.ChapterTypeFree,
						ViewCount:      1000,
					},
				},
				MidChapterList: []mangoplus.Chapter{
					{ChapterID: 5, ChapterType: mangoplus.ChapterTypeStandard},
				},
				LastChapterList: []mangoplus.Chapter{
					{ChapterID: 10, ChapterType: mangoplus.ChapterTypeDeluxe},
				},
			},
		},
		ChapterListV2: []mangoplus.Chapter{
			{ChapterID: 1, ChapterType: mangoplus.ChapterTypeFree},
			{ChapterID: 10, ChapterType: mangoplus.ChapterTypeLockedAfterFreeRead},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Title.GetTitleDetailV3 mismatch (-want +got):\n%s", diff)
	}
}

func TestTitleService_GetTitleDetailV3_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Title.GetTitleDetailV3", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Title.GetTitleDetailV3(ctx, 42)
		return err
	})
}

func TestTitleService_GetTitleDetailV3_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_detailV3", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Title.GetTitleDetailV3(t.Context(), 1); err == nil {
		t.Error("Title.GetTitleDetailV3 returned nil error for empty success body, want error")
	}
}

func TestTitleService_GetTitleDetailV3_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/title_detailV3", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Invalid Parameter", Body: "bad id"},
		})
	})

	_, err := client.Title.GetTitleDetailV3(t.Context(), 0)
	if !errors.Is(err, mangoplus.ErrInvalidParameter) {
		t.Errorf("Title.GetTitleDetailV3 error = %v, want ErrInvalidParameter", err)
	}
}

func TestTitleService_GetTitleDetailV3_Rating(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pb   proto.TitleDetailView_Rating
		want mangoplus.Rating
	}{
		{proto.TitleDetailView_ALLAGE, mangoplus.RatingAllAges},
		{proto.TitleDetailView_TEEN, mangoplus.RatingTeen},
		{proto.TitleDetailView_TEENPLUS, mangoplus.RatingTeenPlus},
		{proto.TitleDetailView_MATURE, mangoplus.RatingMature},
		{proto.TitleDetailView_Rating(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.pb.String(), func(t *testing.T) {
			t.Parallel()
			got := testGetTitleDetailV3(t, &proto.TitleDetailView{Rating: tt.pb})
			if got.Rating != tt.want {
				t.Errorf("Rating is %q, want %q", got.Rating, tt.want)
			}
		})
	}
}

func TestTitleService_GetTitleDetailV3_ReleaseSchedule(t *testing.T) {
	t.Parallel()
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
		{proto.TitleDetailView_ReleaseSchedule(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.pb.String(), func(t *testing.T) {
			t.Parallel()
			got := testGetTitleDetailV3(t, &proto.TitleDetailView{
				TitleLabels: &proto.TitleDetailView_TitleLabels{ReleaseSchedule: tt.pb},
			})
			if got.Labels.ReleaseSchedule != tt.want {
				t.Errorf("Labels.ReleaseSchedule is %q, want %q", got.Labels.ReleaseSchedule, tt.want)
			}
		})
	}
}

func TestTitleService_GetTitleDetailV3_LabelCode(t *testing.T) {
	t.Parallel()
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
		{proto.LabelCodes(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.pb.String(), func(t *testing.T) {
			t.Parallel()
			got := testGetTitleDetailV3(t, &proto.TitleDetailView{Label: &proto.Label{Label: tt.pb}})
			if got.Label.Code != tt.want {
				t.Errorf("Label.Code is %q, want %q", got.Label.Code, tt.want)
			}
		})
	}
}

func TestTitleService_GetTitleDetailV3_ChapterType(t *testing.T) {
	t.Parallel()
	got := testGetTitleDetailV3(t, &proto.TitleDetailView{
		ChapterListV2: []*proto.Chapter{
			{ChapterType: proto.ChapterType_FREE},
			{ChapterType: proto.ChapterType_FREE_FOR_FIRST_TIME},
			{ChapterType: proto.ChapterType_STANDARD},
			{ChapterType: proto.ChapterType_DELUXE},
			{ChapterType: proto.ChapterType_LOCKED_AFTER_FREE_READ},
			{ChapterType: proto.ChapterType(-1)},
		},
	})

	want := []mangoplus.ChapterType{
		mangoplus.ChapterTypeFree,
		mangoplus.ChapterTypeFreeForFirstTime,
		mangoplus.ChapterTypeStandard,
		mangoplus.ChapterTypeDeluxe,
		mangoplus.ChapterTypeLockedAfterFreeRead,
		"",
	}
	var gotTypes []mangoplus.ChapterType
	for _, c := range got.ChapterListV2 {
		gotTypes = append(gotTypes, c.ChapterType)
	}
	if diff := cmp.Diff(want, gotTypes); diff != "" {
		t.Errorf("ChapterListV2 chapter types mismatch (-want +got):\n%s", diff)
	}
}
