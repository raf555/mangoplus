package mangoplus

import (
	"errors"
	"net/url"
	"strings"
)

// redactURLError returns err with the secret query param masked in the URL of
// the [*url.Error] that [http.Client.Do] returns. Other errors pass through unchanged.
func redactURLError(err error) error {
	uerr, ok := errors.AsType[*url.Error](err)
	if !ok {
		return err
	}

	redactedErr := *uerr
	redactedErr.URL = redactURL(uerr.URL)
	return &redactedErr
}

func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "****"
	}

	u.RawQuery = maskSecretQuery(u.RawQuery)
	return u.String()
}

// maskSecretQuery masks the secret value in rawQuery, keeping other params as-is.
// It edits the raw query instead of re-encoding it, since encoding would escape the mask's '*'.
func maskSecretQuery(rawQuery string) string {
	var b strings.Builder
	b.Grow(len(rawQuery))
	first := true
	for p := range strings.SplitSeq(rawQuery, "&") {
		if !first {
			b.WriteByte('&')
		}
		first = false

		if k, v, ok := strings.Cut(p, "="); ok && k == "secret" {
			b.WriteString("secret=")
			b.WriteString(maskSecret(v))
			continue
		}
		b.WriteString(p)
	}
	return b.String()
}

func maskSecret(s string) string {
	const keep = 4
	if len(s) <= 2*keep {
		return "****"
	}
	return "****" + s[len(s)-keep:]
}
