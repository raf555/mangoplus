package mangoplus

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
)

// MangaService wraps Manga related APIs.
type MangaService service

type MangaViewer struct {
	MangaPages     []MangaPage
	CurrentChapter Chapter
	// TODO: There should be another field named Pages to hold all kind of page (ads, banner, etc).

	ChapterID          int
	TitleName          string
	ChapterName        string
	NumberOfComments   int
	IsVerticalOnly     bool
	TitleID            int
	StartFromRight     bool
	IsHorizontalOnly   bool
	TitleLanguage      Language
	AvailableLanguages []MangaAvailableLanguage
}

func mangaViewerFromProto(pb *proto.MangaViewer) MangaViewer {
	currentChapter := Chapter{}
	pages := make([]MangaPage, 0, len(pb.GetPages()))

	for _, p := range pb.GetPages() {
		if mp := p.GetMangaPage(); mp != nil {
			pages = append(pages, mangaPageFromProto(mp))
		}

		// usually only in last page
		if lp := p.GetLastPage(); lp != nil {
			cp := lp.GetCurrentChapter()
			if cp != nil {
				currentChapter = chapterFromProto(cp)
			}
		}
	}

	return MangaViewer{
		MangaPages:         pages,
		CurrentChapter:     currentChapter,
		ChapterID:          int(pb.GetChapterId()),
		TitleName:          pb.GetTitleName(),
		ChapterName:        pb.GetChapterName(),
		NumberOfComments:   int(pb.GetNumberOfComments()),
		IsVerticalOnly:     pb.GetIsVerticalOnly(),
		TitleID:            int(pb.GetTitleId()),
		StartFromRight:     pb.GetStartFromRight(),
		IsHorizontalOnly:   pb.GetIsHorizontalOnly(),
		TitleLanguage:      languageFromMangaPlusLang(pb.GetTitleLanguage()),
		AvailableLanguages: slicex.Map(pb.GetTitleAvailableLanguages(), mangaAvailableLanguageFromProto),
	}
}

type MangaAvailableLanguage struct {
	TitleID  int
	Language Language
}

func mangaAvailableLanguageFromProto(pb *proto.MangaViewer_TitleAvailableLanguages) MangaAvailableLanguage {
	return MangaAvailableLanguage{
		TitleID:  int(pb.GetTitleId()),
		Language: languageFromProto(pb.GetLanguage()),
	}
}

type ViewChapterOptions struct {
	SplitImages  bool
	ImageQuality ImageQuality
}

func DefaultViewChapterOptions() ViewChapterOptions {
	return ViewChapterOptions{
		SplitImages:  false,
		ImageQuality: ImageQualitySuperHigh,
	}
}

func (m *MangaService) ViewChapter(ctx context.Context, chapterID int, opts ViewChapterOptions) (MangaViewer, error) {
	u := m.client.baseURL.JoinPath("/manga_viewer")

	split := "no"
	if opts.SplitImages {
		split = "yes"
	}

	quality := opts.ImageQuality
	if quality == "" {
		quality = ImageQualitySuperHigh
	}

	uParams := url.Values{}
	uParams.Set("chapter_id", strconv.Itoa(chapterID))
	uParams.Set("split", split)
	uParams.Set("img_quality", string(quality))
	u.RawQuery = uParams.Encode()

	req, err := m.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return MangaViewer{}, err
	}

	res, err := m.client.protoDo(req)
	if err != nil {
		return MangaViewer{}, err
	}

	v := res.GetMangaViewer()
	if v == nil {
		return MangaViewer{}, errors.New("mangoplus: unexpected nil manga viewer")
	}

	return mangaViewerFromProto(v), nil
}
