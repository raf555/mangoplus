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

func TestCommentService_ListComments_Success(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t, mangoplus.WithSecret("s"))

	const createdTS = 1_700_000_000

	mux.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{
			"chapter_id": "42",
			"secret":     "s",
			"os":         "android",
			"os_ver":     mangoplus.DefaultOSVersion,
			"app_ver":    mangoplus.DefaultAppVersion,
		})
		testWriteSuccess(t, w, &proto.SuccessResult{
			Data: &proto.SuccessResult_CommentListView{
				CommentListView: &proto.CommentListView{
					Comments: []*proto.Comment{
						{
							Id:            7,
							Index:         1,
							UserName:      "reader",
							IconUrl:       "https://img/icon",
							IsMyComment:   true,
							AlreadyLiked:  true,
							NumberOfLikes: 12,
							Body:          "nice chapter",
							Created:       createdTS,
						},
						{Id: 8, Index: 2},
					},
					IfSetUserName: true,
					TitleDetailComment: &proto.CommentListView_TitleDetailComment{
						TitleId:     100,
						TitleName:   "The Title",
						ChapterName: "#001",
						Credit:      "Someone",
					},
				},
			},
		})
	})

	got, err := client.Comment.ListComments(t.Context(), 42)
	if err != nil {
		t.Fatalf("Comment.ListComments returned error: %v", err)
	}

	want := mangoplus.CommentListView{
		Comments: []mangoplus.Comment{
			{
				ID:            7,
				Index:         1,
				UserName:      "reader",
				IconURL:       "https://img/icon",
				IsMyComment:   true,
				AlreadyLiked:  true,
				NumberOfLikes: 12,
				Body:          "nice chapter",
				CreatedAt:     time.Unix(createdTS, 0),
			},
			{ID: 8, Index: 2},
		},
		TitleDetail: mangoplus.CommentTitleDetail{
			TitleID:     100,
			TitleName:   "The Title",
			ChapterName: "#001",
			Credit:      "Someone",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Comment.ListComments mismatch (-want +got):\n%s", diff)
	}
}

func TestCommentService_ListComments_RequestFailed(t *testing.T) {
	t.Parallel()
	testNewRequestAndDoFailure(t, "Comment.ListComments", func(ctx context.Context, client *mangoplus.Client) error {
		_, err := client.Comment.ListComments(ctx, 42)
		return err
	})
}

func TestCommentService_ListComments_NilView(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		testWriteSuccess(t, w, &proto.SuccessResult{})
	})

	if _, err := client.Comment.ListComments(t.Context(), 1); err == nil {
		t.Error("Comment.ListComments returned nil error for empty success body, want error")
	}
}

func TestCommentService_ListComments_ProtoError(t *testing.T) {
	t.Parallel()
	client, mux, _ := setup(t)

	mux.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		testWriteError(t, w, &proto.ErrorResult{
			EnglishPopup: &proto.Popup_OSDefault{Subject: "Invalid Parameter", Body: "bad id"},
		})
	})

	_, err := client.Comment.ListComments(t.Context(), 0)
	if !errors.Is(err, mangoplus.ErrInvalidParameter) {
		t.Errorf("Comment.ListComments error = %v, want ErrInvalidParameter", err)
	}
}
