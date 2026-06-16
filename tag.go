package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type Tag struct {
	Name string
	Slug string
}

func tagFromProto(pb *proto.Tag) Tag {
	return Tag{
		Name: pb.GetTag(),
		Slug: pb.GetSlug(),
	}
}
