package mangoplus

import "github.com/raf555/mangoplus/internal/proto"

type PlanType string

const (
	PlanTypeBasic    PlanType = "basic"
	PlanTypeStandard PlanType = "standard"
	PlanTypeDeluxe   PlanType = "deluxe"
)

func planTypeFromString(s string) PlanType {
	switch PlanType(s) {
	case PlanTypeDeluxe:
		return PlanTypeDeluxe
	case PlanTypeStandard:
		return PlanTypeStandard
	default:
		return PlanTypeBasic
	}
}

type Subscription struct {
	PlanType PlanType
}

func subscriptionFromProto(pb *proto.Subscription) Subscription {
	return Subscription{
		PlanType: planTypeFromString(pb.GetPlanType()),
	}
}
