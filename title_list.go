package mangoplus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
	"github.com/raf555/mangoplus/internal/timex"
)

// TitleListService wraps Title list related APIs.
type TitleListService service

type AllTitlesViewV2 struct {
	AllTitlesGroup []AllTitlesGroupV2
}

func allTitlesViewV2FromProto(pb *proto.AllTitlesViewV2) AllTitlesViewV2 {
	return AllTitlesViewV2{
		AllTitlesGroup: slicex.Map(pb.GetAllTitlesGroup(), allTitlesGroupV2FromProto),
	}
}

type AllTitlesGroupV2 struct {
	TheTitle                  string
	Titles                    []Title
	Tags                      []Tag
	Label                     Label
	NextChapterStartTimestamp time.Time
}

func allTitlesGroupV2FromProto(pb *proto.AllTitlesGroup) AllTitlesGroupV2 {
	return AllTitlesGroupV2{
		TheTitle:                  pb.GetTheTitle(),
		Titles:                    slicex.Map(pb.GetTitles(), titleFromProto),
		Tags:                      slicex.Map(pb.GetTags(), tagFromProto),
		Label:                     labelFromProto(pb.GetLabel()),
		NextChapterStartTimestamp: timex.Unix(int64(pb.GetNextChapterStartTimestamp())),
	}
}

func (t *TitleListService) AllV2(ctx context.Context) (AllTitlesViewV2, error) {
	u := t.client.baseURL.JoinPath("/title_list/allV2")

	req, err := t.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return AllTitlesViewV2{}, err
	}

	res, err := t.client.protoDo(req)
	if err != nil {
		return AllTitlesViewV2{}, err
	}

	v := res.GetAllTitlesViewV2()
	if v == nil {
		return AllTitlesViewV2{}, errors.New("mangoplus: unexpected nil all titles view")
	}

	return allTitlesViewV2FromProto(v), nil
}

type TitleListType string

const (
	TitleListTypeSerializing TitleListType = "serializing"
	TitleListTypeCompleted   TitleListType = "completed"
	TitleListTypeOneShot     TitleListType = "one-shot"
)

func (t *TitleListService) AllV3(ctx context.Context, typ TitleListType) (SearchView, error) {
	u := t.client.baseURL.JoinPath("/title_list/all_v3")

	uParams := englishParams()
	uParams.Set("type", string(typ))
	u.RawQuery = uParams.Encode()

	return t.searchView(ctx, u)
}

func (t *TitleListService) Search(ctx context.Context, lang Language) (SearchView, error) {
	clang, ok := LanguageToLang(lang)
	if !ok {
		return SearchView{}, fmt.Errorf("mangoplus: unsupported language %q", lang)
	}

	u := t.client.baseURL.JoinPath("/title_list/search")

	uParams := englishParams()
	uParams.Set("clang", string(clang))
	u.RawQuery = uParams.Encode()

	return t.searchView(ctx, u)
}

func (t *TitleListService) searchView(ctx context.Context, u *url.URL) (SearchView, error) {
	req, err := t.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return SearchView{}, err
	}

	res, err := t.client.protoDo(req)
	if err != nil {
		return SearchView{}, err
	}

	v := res.GetSearchView()
	if v == nil {
		return SearchView{}, errors.New("mangoplus: unexpected nil search view")
	}

	return searchViewFromProto(v), nil
}

func englishParams() url.Values {
	return url.Values{
		"lang":  {string(LangEnglish)},
		"clang": {string(LangEnglish)},
	}
}
