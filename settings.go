package mangoplus

import (
	"context"
	"errors"
	"net/http"

	"github.com/raf555/mangoplus/internal/proto"
)

// SettingsService wraps Settings related APIs.
type SettingsService service

type SettingsView struct {
	UserName     string
	IconURL      string
	Subscription Subscription
}

func settingsViewFromProto(pb *proto.SettingsViewV2) SettingsView {
	return SettingsView{
		UserName:     pb.GetUserName(),
		IconURL:      pb.GetMyIcon().GetImageUrl(),
		Subscription: subscriptionFromProto(pb.GetUserSubscription()),
	}
}

func (s *SettingsService) ViewV2(ctx context.Context) (SettingsView, error) {
	u := s.client.baseURL.JoinPath("/settings_v2")

	uParams := englishParams()
	uParams.Set("viewer_mode", "horizontal")
	u.RawQuery = uParams.Encode()

	req, err := s.client.NewRequest(ctx, http.MethodGet, u.String())
	if err != nil {
		return SettingsView{}, err
	}

	res, err := s.client.protoDo(req)
	if err != nil {
		return SettingsView{}, err
	}

	v := res.GetSettingsViewV2()
	if v == nil {
		return SettingsView{}, errors.New("mangoplus: unexpected nil settings view")
	}

	return settingsViewFromProto(v), nil
}
