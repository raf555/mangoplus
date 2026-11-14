package mangoplus

import (
	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
)

type SearchView struct {
	TopBanners     []Banner
	AllTags        []Tag
	AllTitlesGroup []AllTitlesGroupV2
	Contents       []SearchViewContents
}

func searchViewFromProto(pb *proto.SearchView) SearchView {
	return SearchView{
		TopBanners:     slicex.Map(pb.GetTopSearchBanners(), bannerFromProto),
		AllTags:        slicex.Map(pb.GetAllTags(), tagFromProto),
		AllTitlesGroup: slicex.Map(pb.GetAllTitlesGroup(), allTitlesGroupV2FromProto),
		Contents:       slicex.Map(pb.GetContents(), searchViewContentsFromProto),
	}
}

type SearchViewContents struct {
	Banner          *Banner
	TitleList       *TitleList
	ChapterPageList *ChapterPageList
}

func searchViewContentsFromProto(pb *proto.SearchView_Contents) SearchViewContents {
	c := SearchViewContents{}
	if b := pb.GetBanner(); b != nil {
		c.Banner = new(bannerFromProto(b))
	}
	if tl := pb.GetTitleList(); tl != nil {
		c.TitleList = new(titleListFromProto(tl))
	}
	if cpl := pb.GetChapterPagesList(); cpl != nil {
		c.ChapterPageList = new(chapterPageListFromProto(cpl))
	}
	return c
}

type TitleList struct {
	ListName       string
	FeaturedTitles []Title
}

func titleListFromProto(pb *proto.TitleList) TitleList {
	return TitleList{
		ListName:       pb.GetListName(),
		FeaturedTitles: slicex.Map(pb.GetFeaturedTitles(), titleFromProto),
	}
}

type ChapterPageList struct {
	ListName     string
	ChapterPages []ChapterPages
}

func chapterPageListFromProto(pb *proto.ChapterPageList) ChapterPageList {
	return ChapterPageList{
		ListName:     pb.GetListName(),
		ChapterPages: slicex.Map(pb.GetChapterPages(), chapterPagesFromProto),
	}
}

type ChapterPages struct {
	ChapterID        int
	TitleID          int
	Name             string
	Author           string
	FavoriteImageURL string
	Pages            []MangaPage
}

func chapterPagesFromProto(pb *proto.ChapterPages) ChapterPages {
	pages := make([]MangaPage, 0, len(pb.GetPages()))
	for _, p := range pb.GetPages() {
		if mp := p.GetMangaPage(); mp != nil {
			pages = append(pages, mangaPageFromProto(mp))
		}
	}

	return ChapterPages{
		ChapterID:        int(pb.GetChapterId()),
		TitleID:          int(pb.GetTitleId()),
		Name:             pb.GetName(),
		Author:           pb.GetAuthor(),
		FavoriteImageURL: pb.GetFavoriteImageUrl(),
		Pages:            pages,
	}
}
