package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type Label struct {
	Label       LabelCode
	Description string
}

func labelFromProto(pb *proto.Label) Label {
	return Label{
		Label:       labelCodeFromProto(pb.GetLabel()),
		Description: pb.GetDescription(),
	}
}

type LabelCode string

const (
	LabelCodeWJ       LabelCode = "WJ"
	LabelCodeSQ       LabelCode = "SQ"
	LabelCodeVJ       LabelCode = "VJ"
	LabelCodeYJ       LabelCode = "YJ"
	LabelCodeJPlus    LabelCode = "J_PLUS"
	LabelCodeRevival  LabelCode = "REVIVAL"
	LabelCodeCreators LabelCode = "CREATORS"
	LabelCodeMEE      LabelCode = "MEE"
	LabelCodeTYJ      LabelCode = "TYJ"
	LabelCodeOthers   LabelCode = "OTHERS"
	LabelCodeSKJ      LabelCode = "SKJ"
	LabelCodeGiga     LabelCode = "GIGA"
	LabelCodeUJ       LabelCode = "UJ"
	LabelCodeDX       LabelCode = "DX"
)

func labelCodeFromProto(pb proto.LabelCodes) LabelCode {
	val, ok := proto.LabelCodes_name[int32(pb)]
	if !ok {
		return ""
	}
	return LabelCode(val)
}
