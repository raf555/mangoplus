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

// CommentService wraps Comment related APIs.
type CommentService service

type Comment struct {
	ID            int
	Index         int
	UserName      string
	IconURL       string
	IsMyComment   bool
	AlreadyLiked  bool
	NumberOfLikes int
	Body          string
	CreatedAt     time.Time
}

func commentFromProto(pb *proto.Comment) Comment {
	return Comment{
		ID:            int(pb.GetId()),
		Index:         int(pb.GetIndex()),
		UserName:      pb.GetUserName(),
		IconURL:       pb.GetIconUrl(),
		IsMyComment:   pb.GetIsMyComment(),
		AlreadyLiked:  pb.GetAlreadyLiked(),
		NumberOfLikes: int(pb.GetNumberOfLikes()),
		Body:          pb.GetBody(),
		CreatedAt:     timex.Unix(int64(pb.GetCreated())),
	}
}

type CommentListView struct {
	Comments    []Comment
	TitleDetail CommentTitleDetail
}

func commentListViewFromProto(pb *proto.CommentListView) CommentListView {
	return CommentListView{
		Comments:    slicex.Map(pb.GetComments(), commentFromProto),
		TitleDetail: commentTitleDetailFromProto(pb.GetTitleDetailComment()),
	}
}

// CommentTitleDetail describes the title and chapter the comments belong to.
type CommentTitleDetail struct {
	TitleID     int
	TitleName   string
	ChapterName string
	Credit      string
}

func commentTitleDetailFromProto(pb *proto.CommentListView_TitleDetailComment) CommentTitleDetail {
	return CommentTitleDetail{
		TitleID:     int(pb.GetTitleId()),
		TitleName:   pb.GetTitleName(),
		ChapterName: pb.GetChapterName(),
		Credit:      pb.GetCredit(),
	}
}

// ListComments returns the comments posted on the chapter with chapterID.
func (c *CommentService) ListComments(ctx context.Context, chapterID int) (CommentListView, error) {
	u := c.client.baseURL.JoinPath("/comment")

	uParams := url.Values{}
	uParams.Set("chapter_id", strconv.Itoa(chapterID))
	u.RawQuery = uParams.Encode()

	req, err := c.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return CommentListView{}, err
	}

	res, err := c.client.protoDo(req)
	if err != nil {
		return CommentListView{}, err
	}

	v := res.GetCommentListView()
	if v == nil {
		return CommentListView{}, errors.New("mangoplus: unexpected nil comment list view")
	}

	return commentListViewFromProto(v), nil
}
