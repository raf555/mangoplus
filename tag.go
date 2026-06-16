package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type Tag struct {
	Tag  string
	Slug string
}

func tagFromProto(pb *proto.Tag) Tag {
	return Tag{
		Tag:  pb.GetTag(),
		Slug: pb.GetSlug(),
	}
}
