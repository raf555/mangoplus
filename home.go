package mangoplus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
	"github.com/raf555/mangoplus/internal/timex"
)

type HomeService service

type HomeViewV6 struct {
	Sections []HomeSection
}

func homeViewV6FromProto(pb *proto.HomeViewV6) HomeViewV6 {
	return HomeViewV6{
		Sections: slicex.Map(pb.GetSections(), homeSectionFromProto),
	}
}

type HomeSection struct {
	Weekly    *WeeklySection
	Ranking   *RankingSection
	TitleList *TitleList
	Banners   *BannerSection
}

func homeSectionFromProto(pb *proto.HomeViewV6_Sections) HomeSection {
	s := HomeSection{}
	if w := pb.GetWeeklySection(); w != nil {
		s.Weekly = new(weeklySectionFromProto(w))
	}
	if r := pb.GetRankingSection(); r != nil {
		s.Ranking = new(rankingSectionFromProto(r))
	}
	if tl := pb.GetTitleListSection(); tl != nil {
		s.TitleList = new(titleListFromProto(tl.GetTitleList()))
	}
	if b := pb.GetBannerSection(); b != nil {
		s.Banners = new(BannerSection{
			Banners: slicex.Map(b.GetBanners(), bannerFromProto),
		})
	}
	return s
}

type WeeklySection struct {
	Contents []WeeklyContent
}

func weeklySectionFromProto(pb *proto.HomeViewV6_WeeklySection) WeeklySection {
	return WeeklySection{
		Contents: slicex.Map(pb.GetContents(), weeklyContentFromProto),
	}
}

type WeeklyContent struct {
	IsUpdated bool
	UpdatedAt time.Time
	Items     []WeeklyContentItem
}

func weeklyContentFromProto(pb *proto.HomeViewV6_WeeklySection_WeeklyContent) WeeklyContent {
	return WeeklyContent{
		IsUpdated: pb.GetIsUpdated(),
		UpdatedAt: timex.Unix(pb.GetUpdatedTimeStamp()),
		Items:     slicex.Map(pb.GetContentItems(), weeklyContentItemFromProto),
	}
}

type WeeklyContentItem struct {
	PRBanner        *Banner
	MVBanner        *MVBanner
	TitleGroups     []OriginalTitleGroup
	CarouselBanners []Banner
}

func weeklyContentItemFromProto(pb *proto.HomeViewV6_WeeklySection_ContentItem) WeeklyContentItem {
	item := WeeklyContentItem{}
	if b := pb.GetPrBanner(); b != nil {
		item.PRBanner = new(bannerFromProto(b))
	}
	if mv := pb.GetMvBanner(); mv != nil {
		item.MVBanner = new(MVBanner{
			ImageURL:   mv.GetImageUrl(),
			TitleGroup: originalTitleGroupFromProto(mv.GetTitleGroups()),
		})
	}
	if tg := pb.GetTitleGroup(); tg != nil {
		item.TitleGroups = slicex.Map(tg.GetTitleGroups(), originalTitleGroupFromProto)
	}
	if cb := pb.GetCarouselBanners(); cb != nil {
		item.CarouselBanners = slicex.Map(cb.GetBanners(), bannerFromProto)
	}
	return item
}

type MVBanner struct {
	ImageURL   string
	TitleGroup OriginalTitleGroup
}

type OriginalTitleGroup struct {
	TheTitle         string
	ChapterNumber    string
	Titles           []UpdatedTitle
	ViewCount        int
	UpdateStatus     TitleUpdateStatus
	ChapterStartTime time.Time
}

func originalTitleGroupFromProto(pb *proto.OriginalTitleGroup) OriginalTitleGroup {
	return OriginalTitleGroup{
		TheTitle:         pb.GetTheTitle(),
		ChapterNumber:    pb.GetChapterNumber(),
		Titles:           slicex.Map(pb.GetTitles(), updatedTitleFromProto),
		ViewCount:        int(pb.GetViewCount()),
		UpdateStatus:     titleUpdateStatusFromProto(pb.GetTitleUpdateStatus()),
		ChapterStartTime: timex.Unix(int64(pb.GetChapterStartTime())),
	}
}

type UpdatedTitle struct {
	Title           Title
	ChapterID       int
	ChapterName     string
	ChapterSubTitle string
	IsLatest        bool
}

func updatedTitleFromProto(pb *proto.UpdatedTitle) UpdatedTitle {
	return UpdatedTitle{
		Title:           titleFromProto(pb.GetTitle()),
		ChapterID:       int(pb.GetChapterId()),
		ChapterName:     pb.GetChapterName(),
		ChapterSubTitle: pb.GetChapterSubTitle(),
		IsLatest:        pb.GetIsLatest(),
	}
}

type TitleUpdateStatus string

const (
	TitleUpdateStatusNone      TitleUpdateStatus = "NONE"
	TitleUpdateStatusNew       TitleUpdateStatus = "NEW"
	TitleUpdateStatusUp        TitleUpdateStatus = "UP"
	TitleUpdateStatusReedition TitleUpdateStatus = "REEDITION"
	TitleUpdateStatusCreators  TitleUpdateStatus = "CREATORS_STATUS"
	TitleUpdateStatusOurPicks  TitleUpdateStatus = "OUR_PICKS"
)

func titleUpdateStatusFromProto(pb proto.TitleUpdateStatus) TitleUpdateStatus {
	val, ok := proto.TitleUpdateStatus_name[int32(pb)]
	if !ok {
		return ""
	}
	return TitleUpdateStatus(val)
}

type RankingSection struct {
	Tabs []RankingTab
}

func rankingSectionFromProto(pb *proto.HomeViewV6_RankingSection) RankingSection {
	return RankingSection{
		Tabs: slicex.Map(pb.GetRankingTabs(), rankingTabFromProto),
	}
}

type RankingTab struct {
	Type         RankingType
	RankedTitles []TitleRankingGroup
}

func rankingTabFromProto(pb *proto.HomeViewV6_RankingSection_RankingTab) RankingTab {
	return RankingTab{
		Type:         rankingTypeFromProto(pb.GetTabType()),
		RankedTitles: slicex.Map(pb.GetRankedTitles(), titleRankingGroupFromProto),
	}
}

type BannerSection struct {
	Banners []Banner
}

func (h *HomeService) ViewV6(ctx context.Context) (HomeViewV6, error) {
	u := h.client.baseURL.JoinPath("/home_v6")

	uParams := englishParams()
	uParams.Set("viewer_mode", "horizontal")
	u.RawQuery = uParams.Encode()

	req, err := h.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return HomeViewV6{}, err
	}

	res, err := h.client.protoDo(req)
	if err != nil {
		return HomeViewV6{}, err
	}

	v := res.GetHomeViewV6()
	if v == nil {
		return HomeViewV6{}, errors.New("mangoplus: unexpected nil home view")
	}

	return homeViewV6FromProto(v), nil
}

type HomePreviewViewV2 struct {
	Sections []PreviewSection
}

func homePreviewViewV2FromProto(pb *proto.HomePreviewViewV2) HomePreviewViewV2 {
	return HomePreviewViewV2{
		Sections: slicex.Map(pb.GetPreviewSections(), previewSectionFromProto),
	}
}

type PreviewSection struct {
	ID   int
	Name string
	Type PreviewSectionType
	Tabs []PreviewTab
}

func previewSectionFromProto(pb *proto.HomePreviewViewV2_PreviewSection) PreviewSection {
	return PreviewSection{
		ID:   int(pb.GetSectionId()),
		Name: pb.GetSectionName(),
		Type: previewSectionTypeFromProto(pb.GetSectionType()),
		Tabs: slicex.Map(pb.GetPreviewTabs(), previewTabFromProto),
	}
}

type PreviewSectionType string

const (
	PreviewSectionTypeUpdate  PreviewSectionType = "UPDATE"
	PreviewSectionTypeChapter PreviewSectionType = "CHAPTER"
)

func previewSectionTypeFromProto(pb proto.HomePreviewViewV2_PreviewSection_SectionType) PreviewSectionType {
	val, ok := proto.HomePreviewViewV2_PreviewSection_SectionType_name[int32(pb)]
	if !ok {
		return ""
	}
	return PreviewSectionType(val)
}

type PreviewTab struct {
	Language Language
	Titles   []PreviewTitle
}

func previewTabFromProto(pb *proto.HomePreviewViewV2_PreviewTab) PreviewTab {
	return PreviewTab{
		Language: languageFromProto(pb.GetLanguage()),
		Titles:   slicex.Map(pb.GetTitles(), previewTitleFromProto),
	}
}

type PreviewTitle struct {
	ChapterID    int
	TitleName    string
	Author       string
	ThumbnailURL string
	Pages        []PreviewPage
}

func previewTitleFromProto(pb *proto.HomePreviewViewV2_PreviewTitle) PreviewTitle {
	return PreviewTitle{
		ChapterID:    int(pb.GetChapterId()),
		TitleName:    pb.GetTitleName(),
		Author:       pb.GetAuthor(),
		ThumbnailURL: pb.GetThumbnailUrl(),
		Pages:        slicex.Map(pb.GetPages(), previewPageFromProto),
	}
}

type PreviewPage struct {
	ImageURL string
	Width    int
	Height   int
}

func previewPageFromProto(pb *proto.HomePreviewViewV2_PreviewPage) PreviewPage {
	return PreviewPage{
		ImageURL: pb.GetImageUrl(),
		Width:    int(pb.GetWidth()),
		Height:   int(pb.GetHeight()),
	}
}

func (h *HomeService) PreviewV2(ctx context.Context, lang Language) (HomePreviewViewV2, error) {
	clang, ok := LanguageToLang(lang)
	if !ok {
		return HomePreviewViewV2{}, fmt.Errorf("mangoplus: unsupported language %q", lang)
	}

	u := h.client.baseURL.JoinPath("/preview_v2")

	uParams := englishParams()
	uParams.Set("clang", string(clang))
	uParams.Set("viewer_mode", "horizontal")
	u.RawQuery = uParams.Encode()

	req, err := h.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return HomePreviewViewV2{}, err
	}

	res, err := h.client.protoDo(req)
	if err != nil {
		return HomePreviewViewV2{}, err
	}

	v := res.GetHomePreviewViewV2()
	if v == nil {
		return HomePreviewViewV2{}, errors.New("mangoplus: unexpected nil home preview view")
	}

	return homePreviewViewV2FromProto(v), nil
}
