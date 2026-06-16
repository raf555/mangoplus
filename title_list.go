package mangoplus

import (
	"context"
	"net/http"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/slicex"
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
	TheTitle string
	Titles   []TitleV2
}

func allTitlesGroupV2FromProto(pb *proto.AllTitlesGroup) AllTitlesGroupV2 {
	return AllTitlesGroupV2{
		TheTitle: pb.GetTheTitle(),
		Titles:   slicex.Map(pb.GetTitles(), titleV2FromProto),
	}
}

type TitleV2 struct {
	TitleID           int
	Name              string
	Author            string
	PortraitImageURL  string
	LandscapeImageURL string
	Language          Language
}

func titleV2FromProto(pb *proto.Title) TitleV2 {
	return TitleV2{
		TitleID:           int(pb.GetTitleId()),
		Name:              pb.GetName(),
		Author:            pb.GetAuthor(),
		PortraitImageURL:  pb.GetPortraitImageUrl(),
		LandscapeImageURL: pb.GetLandscapeImageUrl(),
		Language:          languageFromProto(pb.GetLanguage()),
	}
}

func (t *TitleListService) AllV2(ctx context.Context) (AllTitlesViewV2, error) {
	u := t.client.baseURL.JoinPath("/title_list/allV2")

	req, err := t.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return AllTitlesViewV2{}, err
	}

	res, err := t.client.do(req)
	if err != nil {
		return AllTitlesViewV2{}, err
	}

	return allTitlesViewV2FromProto(res.GetAllTitlesViewV2()), nil
}
