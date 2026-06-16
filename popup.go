package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type PopupOsDefault struct {
	Subject  string
	Body     string
	Language Language
}

func popupOsDefaultFromProto(pb *proto.Popup_OSDefault) PopupOsDefault {
	return PopupOsDefault{
		Subject:  pb.GetSubject(),
		Body:     pb.GetBody(),
		Language: languageFromProto(pb.GetLanguage()),
	}
}
