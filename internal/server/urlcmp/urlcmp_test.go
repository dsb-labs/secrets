package urlcmp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidsbond/keeper/internal/server/urlcmp"
)

func TestHostKey(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name: "empty",
		},
		{
			Name:  "whitespace only",
			Input: "   \t\n",
		},
		{
			Name:     "schemeless bare domain",
			Input:    "facebook.com",
			Expected: "facebook.com",
		},
		{
			Name:     "schemeless with path",
			Input:    "example.com/some/path?x=1#frag",
			Expected: "example.com",
		},
		{
			Name:     "with scheme and path",
			Input:    "https://Accounts.Google.com/login",
			Expected: "accounts.google.com",
		},
		{
			Name:     "strip port (schemeless)",
			Input:    "google.com:8443",
			Expected: "google.com",
		},
		{
			Name:     "strip port (scheme)",
			Input:    "https://example.com:443/a/b",
			Expected: "example.com",
		},
		{
			Name:     "lowercase normalization",
			Input:    "HTTPS://EXAMPLE.COM",
			Expected: "example.com",
		},
		{
			Name:     "trailing dot removed",
			Input:    "example.com.",
			Expected: "example.com",
		},
		{
			Name:     "unicode IDN normalized to punycode",
			Input:    "пример.рф",
			Expected: "xn--e1afmkfd.xn--p1ai",
		},
		{
			Name:  "invalid url yields not ok",
			Input: "http://[::1",
		},
		{
			Name:  "no host after parse yields not ok",
			Input: "https:///just/path",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			actual, ok := urlcmp.HostKey(tc.Input)
			if tc.Expected != "" {
				assert.True(t, ok)
			}

			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}

func TestSiteKey(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name: "empty",
		},
		{
			Name:     "google and accounts share site",
			Input:    "https://accounts.google.com/login",
			Expected: "google.com",
		},
		{
			Name:     "bare google is same site",
			Input:    "google.com",
			Expected: "google.com",
		},
		{
			Name:     "subdomain collapses to registrable domain",
			Input:    "mail.google.com",
			Expected: "google.com",
		},
		{
			Name:     "psl behavior: github.io registrable includes subdomain",
			Input:    "foo.github.io",
			Expected: "foo.github.io",
		},
		{
			Name:     "ip address returned as-is (port stripped by HostKey first)",
			Input:    "127.0.0.1:8080",
			Expected: "127.0.0.1",
		},
		{
			Name:     "localhost returned as-is",
			Input:    "localhost:3000",
			Expected: "localhost",
		},
		{
			Name:     "unicode IDN site key uses punycode",
			Input:    "https://пример.рф/path",
			Expected: "xn--e1afmkfd.xn--p1ai",
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			actual, ok := urlcmp.SiteKey(tc.Input)
			if tc.Expected != "" {
				assert.True(t, ok)
			}

			assert.EqualValues(t, tc.Expected, actual)
		})
	}
}
