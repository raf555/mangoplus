package mangoplus

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/raf555/mangoplus/internal/encodingx"
	"github.com/raf555/mangoplus/internal/proto"
	"github.com/raf555/mangoplus/internal/xrand"
	"golang.org/x/net/publicsuffix"

	protopkg "google.golang.org/protobuf/proto"
)

// Client is a wrapper for MangaPlus APIs.
//
// Note: Services do not guarantee all fields are populated, some may be zero values.
// Callers are responsible for handling this appropriately.
type Client struct {
	httpClient *http.Client

	baseURL *url.URL

	userAgent string

	osVersion  string
	appVersion string

	androidID string
	secret    string

	common service

	// Services

	Registration *RegistrationService
	TitleList    *TitleListService
	Title        *TitleService
}

type service struct {
	client *Client
}

// NewClient returns a new [Client] configured with the provided [ClientOptionsFunc].
// Default configuration of [Client] is ready to use after [Client.Register] call is successful.
//
// It is recommended to provide secret via [WithSecret] option if you already have it / have already registered it previously.
func NewClient(opts ...ClientOptionsFunc) (*Client, error) {
	co := &clientOptions{}
	for _, opt := range opts {
		if err := opt(co); err != nil {
			return nil, err
		}
	}

	return newClient(co)
}

func newClient(opts *clientOptions) (*Client, error) {
	var err error

	c := &Client{}

	if opts.httpClient != nil {
		c.httpClient = opts.httpClient
	} else {
		c.httpClient = &http.Client{}
	}

	if opts.transport != nil {
		c.httpClient.Transport = opts.transport
	}

	if opts.timeout != nil {
		c.httpClient.Timeout = *opts.timeout
	}

	if opts.cookieJar != nil {
		c.httpClient.Jar = opts.cookieJar
	} else {
		toples, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}
		c.httpClient.Jar = toples
	}

	if opts.userAgent != "" {
		c.userAgent = opts.userAgent
	} else {
		c.userAgent = defaultUserAgent
	}

	if opts.osVersion != "" {
		c.osVersion = opts.osVersion
	} else {
		c.osVersion = defaultOSVersion
	}

	if opts.appVersion != "" {
		c.appVersion = opts.appVersion
	} else {
		c.appVersion = defaultAppVersion
	}

	if opts.secret != "" && opts.androidID != "" {
		return nil, errors.New("mangoplus: both secret and androidID can't be provided at the same time")
	}

	if opts.secret != "" {
		c.secret = opts.secret
	} else if opts.androidID != "" {
		c.androidID = opts.androidID
	} else {
		c.androidID, err = xrand.GenerateAndroidID()
		if err != nil {
			return nil, err
		}
	}

	// TODO: maybe base URL option?
	c.baseURL, err = url.Parse(appAPIURL)
	if err != nil {
		return nil, err
	}

	c.common.client = c

	// services
	c.Registration = (*RegistrationService)(&c.common)
	c.TitleList = (*TitleListService)(&c.common)
	c.Title = (*TitleService)(&c.common)

	if opts.autoRegister {
		_, err = c.Register(opts.registCtx)
		if err != nil {
			return nil, fmt.Errorf("mangoplus: auto register: %w", err)
		}
	}

	return c, nil
}

// CookieJar returns [http.CookieJar] used by [Client].
func (c *Client) CookieJar() http.CookieJar {
	return c.httpClient.Jar
}

// Secret returns secret key used by [Client] for communicating with MangaPlus API.
//
// It may be empty if not registered yet.
func (c *Client) Secret() string {
	return c.secret
}

// Register registers current [Client] to MangaPlus.
// It also sets and returns [Client]'s secret when successfully called.
//
// Register is noop when secret key is already set.
func (c *Client) Register(ctx context.Context) (string, error) {
	if c.secret != "" {
		return c.secret, nil
	}

	deviceToken := encodingx.MD5Hex(c.androidID)
	securityKey := encodingx.MD5Hex(deviceToken + "4Kin9vGg")

	res, err := c.Registration.Register(ctx, deviceToken, securityKey)
	if err != nil {
		return "", err
	}

	c.secret = res.DeviceSecret
	return res.DeviceSecret, nil
}

type requestOptions struct {
	params  url.Values
	headers http.Header

	bodyContentType string
	body            io.Reader
}

type RequestOptionsFunc func(*requestOptions) error

func WithRequestParams(u url.Values) RequestOptionsFunc {
	return func(ro *requestOptions) error {
		for k, vs := range u {
			for _, v := range vs {
				ro.params.Add(k, v)
			}
		}

		return nil
	}
}

func WithRequestHeaders(h http.Header) RequestOptionsFunc {
	return func(ro *requestOptions) error {
		for k, vs := range h {
			for _, v := range vs {
				ro.headers.Add(k, v)
			}
		}

		return nil
	}
}

func WithRequestBody(contentType string, b io.Reader) RequestOptionsFunc {
	return func(ro *requestOptions) error {
		ro.bodyContentType = contentType
		ro.body = b
		return nil
	}
}

// NewRequest creates a new [http.Request] with necessary information attached to it.
// Additional data may be provided via [RequestOptionsFunc] if needed.
//
// If u has prefix /, it is assumed to be a URL path and part of the [Client] base URL. Otherwise, it'll be treated as-is.
func (c *Client) NewRequest(ctx context.Context, method string, u string, opts ...RequestOptionsFunc) (*http.Request, error) {
	if c.osVersion == "" {
		return nil, errors.New("mangoplus: empty client os version")
	}

	if c.appVersion == "" {
		return nil, errors.New("mangoplus: empty client app version")
	}

	var (
		uri *url.URL
		err error
	)

	if strings.HasPrefix(u, "/") {
		uri, err = c.baseURL.Parse(u)
	} else {
		uri, err = url.Parse(u)
	}

	if err != nil {
		return nil, err
	}

	reqOpts := &requestOptions{
		params:  uri.Query(),
		headers: make(http.Header),
	}

	for _, opt := range opts {
		err = opt(reqOpts)
		if err != nil {
			return nil, err
		}
	}

	// overwrite the rest with ours
	reqOpts.params.Set("os", "android")
	reqOpts.params.Set("os_ver", c.osVersion)
	reqOpts.params.Set("app_ver", c.appVersion)
	// to force the API to return JSON response. Right now we use protobuf since it should be faster.
	// reqOpts.params.Set("format", "json")
	if c.secret != "" {
		reqOpts.params.Set("secret", c.secret)
	}

	reqOpts.headers.Set("User-Agent", c.userAgent)

	uri.RawQuery = reqOpts.params.Encode()

	var body io.Reader
	if reqOpts.body != nil {
		body = reqOpts.body
		reqOpts.headers.Set("Content-Type", reqOpts.bodyContentType)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	maps.Copy(req.Header, reqOpts.headers)

	return req, nil
}

func (c *Client) preDo(req *http.Request) (*proto.Response, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, &APIError{
			StatusCode: res.StatusCode,
			Headers:    res.Header,
			RawBody:    b,
		}
	}

	var resPb proto.Response
	err = protopkg.Unmarshal(b, &resPb)
	if err != nil {
		return nil, err
	}

	return &resPb, nil
}

func (c *Client) do(req *http.Request) (*proto.SuccessResult, error) {
	resPb, err := c.preDo(req)
	if err != nil {
		return nil, err
	}

	errPb := resPb.GetError()
	if errPb != nil {
		return nil, protoErrorFromProto(errPb)
	}

	return resPb.GetSuccess(), nil
}
