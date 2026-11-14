package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type Banner struct {
	ID        int
	ImageURL  string
	Width     int
	Height    int
	ActionURL string
}

func bannerFromProto(pb *proto.Banner) Banner {
	return Banner{
		ID:        int(pb.GetId()),
		ImageURL:  pb.GetImageUrl(),
		Width:     int(pb.GetWidth()),
		Height:    int(pb.GetHeight()),
		ActionURL: pb.GetAction().GetUrl(),
	}
}
