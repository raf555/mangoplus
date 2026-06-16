package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type MangaPage struct {
	ImageURL string
	Width    int
	Height   int
	Type     PageType
}

func mangaPageFromProto(pb *proto.Page_MangaPage) MangaPage {
	return MangaPage{
		ImageURL: pb.GetImageUrl(),
		Width:    int(pb.GetWidth()),
		Height:   int(pb.GetHeight()),
		Type:     pageTypeFromProto(pb.GetType()),
	}
}

type PageType string

const (
	PageTypeSingle PageType = "SINGLE"
	PageTypeLeft   PageType = "LEFT"
	PageTypeRight  PageType = "RIGHT"
	PageTypeDouble PageType = "DOUBLE"
)

func pageTypeFromProto(pb proto.Page_PageType) PageType {
	val, ok := proto.Page_PageType_name[int32(pb)]
	if !ok {
		return ""
	}
	return PageType(val)
}

type ImageQuality string

const (
	ImageQualityLow       ImageQuality = "low"
	ImageQualityHigh      ImageQuality = "high"
	ImageQualitySuperHigh ImageQuality = "super_high"
)
