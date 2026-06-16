package mangoplus

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
	"github.com/raf555/mangoplus/internal/timex"
)

// TitleService wraps Title related APIs.
type TitleService service

type Title struct {
	TitleID           int
	Name              string
	Author            string
	PortraitImageURL  string
	LandscapeImageURL string
	Language          Language
}

func titleFromProto(pb *proto.Title) Title {
	return Title{
		TitleID:           int(pb.GetTitleId()),
		Name:              pb.GetName(),
		Author:            pb.GetAuthor(),
		PortraitImageURL:  pb.GetPortraitImageUrl(),
		LandscapeImageURL: pb.GetLandscapeImageUrl(),
		Language:          languageFromProto(pb.GetLanguage()),
	}
}

type TitleDetailView struct {
	Title              Title
	ImageURL           string
	Overview           string
	BackgroundImageURL string
	NextTimestamp      time.Time
	IsSimulRelease     bool
	Rating             Rating
	NumberOfViews      int
	Languages          []TitleLanguage
	Tags               []Tag
	Labels             TitleLabels
	Label              Label
	ChapterListGroup   []ChapterGroup
	ChapterListV2      []Chapter
}

func titleDetailViewFromProto(pb *proto.TitleDetailView) TitleDetailView {
	return TitleDetailView{
		Title:              titleFromProto(pb.GetTitle()),
		ImageURL:           pb.GetTitleImageUrl(),
		Overview:           pb.GetOverview(),
		BackgroundImageURL: pb.GetBackgroundImageUrl(),
		NextTimestamp:      timex.Unix(int64(pb.GetNextTimeStamp())),
		IsSimulRelease:     pb.GetIsSimulReleased(),
		Rating:             ratingFromProto(pb.GetRating()),
		NumberOfViews:      int(pb.GetNumberOfViews()),
		Languages:          slicex.Map(pb.GetTitleLanguages(), titleLanguageFromProto),
		Tags:               slicex.Map(pb.GetTags(), tagFromProto),
		Labels:             titleLabelsFromProto(pb.GetTitleLabels()),
		Label:              labelFromProto(pb.GetLabel()),
		ChapterListGroup:   slicex.Map(pb.GetChapterListGroup(), chapterGroupFromProto),
		ChapterListV2:      slicex.Map(pb.GetChapterListV2(), chapterFromProto),
	}
}

type Rating string

const (
	RatingAllAges  Rating = "ALLAGES"
	RatingTeen     Rating = "TEEN"
	RatingTeenPlus Rating = "TEENPLUS"
	RatingMature   Rating = "MATURE"
)

func ratingFromProto(pb proto.TitleDetailView_Rating) Rating {
	val, ok := proto.TitleDetailView_Rating_name[int32(pb)]
	if !ok {
		return ""
	}
	return Rating(val)
}

type TitleLanguage struct {
	TitleID  int
	Language Language
}

func titleLanguageFromProto(pb *proto.TitleDetailView_TitleLanguages) TitleLanguage {
	return TitleLanguage{
		TitleID:  int(pb.GetTitleId()),
		Language: languageFromProto(pb.GetLanguage()),
	}
}

type TitleLabels struct {
	ReleaseSchedule ReleaseSchedule
	IsSimulpub      bool
}

func titleLabelsFromProto(pb *proto.TitleDetailView_TitleLabels) TitleLabels {
	return TitleLabels{
		ReleaseSchedule: releaseScheduleFromProto(pb.GetReleaseSchedule()),
		IsSimulpub:      pb.GetIsSimulpub(),
	}
}

type ReleaseSchedule string

const (
	ReleaseScheduleDisabled   ReleaseSchedule = "DISABLED"
	ReleaseScheduleEveryday   ReleaseSchedule = "EVERYDAY"
	ReleaseScheduleWeekly     ReleaseSchedule = "WEEKLY"
	ReleaseScheduleBiweekly   ReleaseSchedule = "BIWEEKLY"
	ReleaseScheduleMonthly    ReleaseSchedule = "MONTHLY"
	ReleaseScheduleBiMonthly  ReleaseSchedule = "BIMONTHLY"
	ReleaseScheduleTriMonthly ReleaseSchedule = "TRIMONTHLY"
	ReleaseScheduleOther      ReleaseSchedule = "OTHER"
	ReleaseScheduleCompleted  ReleaseSchedule = "COMPLETED"
	ReleaseScheduleOneShot    ReleaseSchedule = "ONE_SHOT"
	ReleaseScheduleHiatus     ReleaseSchedule = "HIATUS"
)

func releaseScheduleFromProto(pb proto.TitleDetailView_ReleaseSchedule) ReleaseSchedule {
	val, ok := proto.TitleDetailView_ReleaseSchedule_name[int32(pb)]
	if !ok {
		return ""
	}
	return ReleaseSchedule(val)
}

type ChapterGroup struct {
	ChapterNumbers   string
	FirstChapterList []Chapter
	MidChapterList   []Chapter
	LastChapterList  []Chapter
}

func chapterGroupFromProto(pb *proto.TitleDetailView_ChapterGroup) ChapterGroup {
	return ChapterGroup{
		ChapterNumbers:   pb.GetChapterNumbers(),
		FirstChapterList: slicex.Map(pb.GetFirstChapterList(), chapterFromProto),
		MidChapterList:   slicex.Map(pb.GetMidChapterList(), chapterFromProto),
		LastChapterList:  slicex.Map(pb.GetLastChapterList(), chapterFromProto),
	}
}

func (t *TitleService) GetTitleDetailV3(ctx context.Context, titleID int) (TitleDetailView, error) {
	u := t.client.baseURL.JoinPath("/title_detailV3")

	uParams := url.Values{}
	uParams.Set("title_id", strconv.Itoa(titleID))
	u.RawQuery = uParams.Encode()

	req, err := t.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return TitleDetailView{}, err
	}

	res, err := t.client.protoDo(req)
	if err != nil {
		return TitleDetailView{}, err
	}

	v := res.GetTitleDetailView()
	if v == nil {
		return TitleDetailView{}, errors.New("mangoplus: unexpected nil title detail view")
	}

	return titleDetailViewFromProto(v), nil
}
