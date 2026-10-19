package mangoplus

import "net/url"

// Client defaults, exported for black-box tests.
const (
	DefaultUserAgent  = defaultUserAgent
	DefaultOSVersion  = defaultOSVersion
	DefaultAppVersion = defaultAppVersion
	AppAPIURL         = appAPIURL
)

// SetBaseURL points c at u, so black-box tests can talk to a test server.
func SetBaseURL(c *Client, u *url.URL) {
	c.baseURL = u
}
