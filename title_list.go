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
	Titles   []Title
}

func allTitlesGroupV2FromProto(pb *proto.AllTitlesGroup) AllTitlesGroupV2 {
	return AllTitlesGroupV2{
		TheTitle: pb.GetTheTitle(),
		Titles:   slicex.Map(pb.GetTitles(), titleFromProto),
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
