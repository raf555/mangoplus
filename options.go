package mangoplus

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	defaultUserAgent  = "okhttp/4.12.0"
	defaultOSVersion  = "30"
	defaultAppVersion = "251"

	// The default API URL, which is used by app.
	appAPIURL = "https://jumpg-api.tokyo-cdn.com/api"
	// Browser API URL. Limited to what's available on the browser.
	// webAPIURL = "https://jumpg-webapi.tokyo-cdn.com/api"
)

type clientOptions struct {
	httpClient *http.Client
	transport  http.RoundTripper
	timeout    *time.Duration
	cookieJar  http.CookieJar

	userAgent  string
	osVersion  string
	appVersion string

	androidID string
	secret    string
	// ^ one of those needs to be present. AndroidID for registration if secret is not provided.

	autoRegister bool
	registCtx    context.Context
}

// ClientOptionsFunc is a functional option for providing configuration options to a [Client].
type ClientOptionsFunc func(*clientOptions) error

// WithHTTPClient returns a [ClientOptionsFunc] that sets the [http.Client]
// for a [Client]. If not set, a default [http.Client] will be used.
func WithHTTPClient(httpClient *http.Client) ClientOptionsFunc {
	return func(co *clientOptions) error {
		if httpClient == nil {
			return errors.New("mangoplus: http client must not be nil")
		}

		httpClient := *httpClient
		co.httpClient = &httpClient
		return nil
	}
}

// WithTransport returns a [ClientOptionsFunc] that sets the [http.RoundTripper]
// for a [Client]. This overrides the transport set by [WithHTTPClient]. If not
// set and no HTTP client is provided, the default http.RoundTripper will be used.
func WithTransport(transport http.RoundTripper) ClientOptionsFunc {
	return func(co *clientOptions) error {
		if transport == nil {
			return errors.New("mangoplus: transport must not be nil")
		}

		co.transport = transport
		return nil
	}
}

// WithTimeout returns a [ClientOptionsFunc] that sets the timeout for a [Client].
// This overrides the timeout set by [WithHTTPClient]. If not set and no HTTP
// client is provided, the default [http.Client] with no timeout will be used.
func WithTimeout(timeout time.Duration) ClientOptionsFunc {
	return func(co *clientOptions) error {
		if timeout < 0 {
			return errors.New("mangoplus: timeout must not be negative")
		}

		co.timeout = &timeout
		return nil
	}
}

// WithCookieJar returns a [ClientOptionsFunc] that sets the [http.CookieJar] for a [Client].
// This overrides the cookie jar set by [WithHTTPClient]. If not set and no HTTP
// client is provided, the [Client] will use its own [http.CookieJar]
// which then is retriveable from [Client.CookieJar].
func WithCookieJar(cookieJar http.CookieJar) ClientOptionsFunc {
	return func(co *clientOptions) error {
		if cookieJar == nil {
			return errors.New("mangoplus: cookieJar must not be nil")
		}

		co.cookieJar = cookieJar
		return nil
	}
}

// WithUserAgent returns a [ClientOptionsFunc] that sets the User-Agent header
// for a [Client]. If not set, a default User-Agent will be used.
func WithUserAgent(userAgent string) ClientOptionsFunc {
	return func(o *clientOptions) error {
		o.userAgent = userAgent
		return nil
	}
}

// WithOSVersion returns a [ClientOptionsFunc] that sets the OS version
// for a [Client]. If not set, a default OS version will be used.
func WithOSVersion(osVer string) ClientOptionsFunc {
	return func(co *clientOptions) error {
		co.osVersion = osVer
		return nil
	}
}

// WithAppVersion returns a [ClientOptionsFunc] that sets the app version
// for a [Client]. If not set, a default app version will be used.
func WithAppVersion(appVer string) ClientOptionsFunc {
	return func(co *clientOptions) error {
		co.appVersion = appVer
		return nil
	}
}

// WithAndroidID returns a [ClientOptionsFunc] that sets the android ID
// for a [Client]. If not set, a random android ID will be generated.
//
// Android ID is used for registering the client to retrieve the MangaPlus secret.
// If secret is already provided, [NewClient] may return error.
func WithAndroidID(androidID string) ClientOptionsFunc {
	return func(co *clientOptions) error {
		co.androidID = androidID
		return nil
	}
}

// WithSecret returns a [ClientOptionsFunc] that sets the secret
// for a [Client]. If not set, android ID will be used for registration.
//
// Secret is used for communicating with the MangaPlus APIs.
// If Android ID is already provided, [NewClient] may return error.
func WithSecret(secret string) ClientOptionsFunc {
	return func(co *clientOptions) error {
		co.secret = secret
		return nil
	}
}

// WithAutoRegister returns a [ClientOptionsFunc] that sets auto registeration for [Client].
// This option is noop when secret is provided.
//
// For example:
//
//	client, err := NewClient()
//	_, err = client.Register(context.Background())
//
// is equivalent to
//
//	client, err := NewClient(WithAutoRegister(context.Background()))
func WithAutoRegister(ctx context.Context) ClientOptionsFunc {
	return func(co *clientOptions) error {
		if ctx == nil {
			return errors.New("mangoplus: ctx must not be nil")
		}

		co.autoRegister = true
		co.registCtx = ctx
		return nil
	}
}
