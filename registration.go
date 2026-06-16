package mangoplus

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

type RegistrationService service

type RegisterResult struct {
	DeviceSecret string
}

// Register registers provided device token and security key to MangaPlus. It returns a device secret usable for most MangaPlus API.
func (r *RegistrationService) Register(ctx context.Context, deviceToken, securityKey string) (RegisterResult, error) {
	u := r.client.baseURL.JoinPath("/register")

	uParams := url.Values{}
	uParams.Set("device_token", deviceToken)
	uParams.Set("security_key", securityKey)

	req, err := r.client.NewRequest(ctx, http.MethodPut, u.String(), WithRequestParams(uParams))
	if err != nil {
		return RegisterResult{}, err
	}

	res, err := r.client.protoDo(req)
	if err != nil {
		return RegisterResult{}, err
	}

	secret := res.GetRegisterationData().GetDeviceSecret()
	if secret == "" {
		return RegisterResult{}, errors.New("mangoplus: unexpected empty secret returned")
	}

	return RegisterResult{
		DeviceSecret: secret,
	}, nil
}
