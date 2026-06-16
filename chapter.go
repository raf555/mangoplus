package mangoplus

import (
	"time"

	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/timex"
)

type ChapterType string

const (
	ChapterTypeFree               ChapterType = "FREE"
	ChapterTypeFreeForFirstTime   ChapterType = "FREE_FOR_FIRST_TIME"
	ChapterTypeStandard           ChapterType = "STANDARD"
	ChapterTypeDeluxe             ChapterType = "DELUXE"
	ChapterTypeLockedAfterFreeRead ChapterType = "LOCKED_AFTER_FREE_READ"
)

func chapterTypeFromProto(pb proto.ChapterType) ChapterType {
	val, ok := proto.ChapterType_name[int32(pb)]
	if !ok {
		return ""
	}
	return ChapterType(val)
}

type Chapter struct {
	TitleID              int
	ChapterID            int
	Name                 string
	SubTitle             string
	ThumbnailURL         string
	StartTimestamp       time.Time
	EndTimestamp         time.Time
	AlreadyViewed        bool
	IsVerticalOnly       bool
	ChapterTicketEndtime time.Time
	ViewedForFree        bool
	IsHorizontalOnly     bool
	ViewCount            int
	CommentCount         int
	IsUpdated            bool
	ChapterType          ChapterType
}

func chapterFromProto(pb *proto.Chapter) Chapter {
	return Chapter{
		TitleID:              int(pb.GetTitleId()),
		ChapterID:            int(pb.GetChapterId()),
		Name:                 pb.GetName(),
		SubTitle:             pb.GetSubTitle(),
		ThumbnailURL:         pb.GetThumbnailUrl(),
		StartTimestamp:       timex.Unix(int64(pb.GetStartTimeStamp())),
		EndTimestamp:         timex.Unix(int64(pb.GetEndTimeStamp())),
		AlreadyViewed:        pb.GetAlreadyViewed(),
		IsVerticalOnly:       pb.GetIsVerticalOnly(),
		ChapterTicketEndtime: timex.Unix(int64(pb.GetChapterTicketEndtime())),
		ViewedForFree:        pb.GetViewedForFree(),
		IsHorizontalOnly:     pb.GetIsHorizontalOnly(),
		ViewCount:            int(pb.GetViewCount()),
		CommentCount:         int(pb.GetCommentCount()),
		IsUpdated:            pb.GetIsUpdated(),
		ChapterType:          chapterTypeFromProto(pb.GetChapterType()),
	}
}
