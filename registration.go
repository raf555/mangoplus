package mangoplus

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// RegistrationService wraps registration APIs.
type RegistrationService service

type RegistrationData struct {
	DeviceSecret string
}

// Register registers provided device token and security key to MangaPlus. It returns a device secret usable for most MangaPlus API.
func (r *RegistrationService) Register(ctx context.Context, deviceToken, securityKey string) (RegistrationData, error) {
	u := r.client.baseURL.JoinPath("/register")

	uParams := url.Values{}
	uParams.Set("device_token", deviceToken)
	uParams.Set("security_key", securityKey)
	u.RawQuery = uParams.Encode()

	req, err := r.client.NewRequest(ctx, http.MethodPut, u.String())
	if err != nil {
		return RegistrationData{}, err
	}

	q := req.URL.Query()
	q.Del("secret")
	req.URL.RawQuery = q.Encode()

	res, err := r.client.protoDo(req)
	if err != nil {
		return RegistrationData{}, err
	}

	secret := res.GetRegisterationData().GetDeviceSecret()
	if secret == "" {
		return RegistrationData{}, errors.New("mangoplus: unexpected empty secret returned")
	}

	return RegistrationData{
		DeviceSecret: secret,
	}, nil
}
