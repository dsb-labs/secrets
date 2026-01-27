// Package urlcmp provides functions that assist in the comparison of urls. This is primarily used for determining
// which logins the server should recommend for which domains.
package urlcmp

import (
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// HostKey extracts and normalizes the hostname from a URL-like string.
//
// It is intended for stable host comparisons and ignores any URL components
// other than the host.
//
// Normalization rules:
//   - If no scheme is present, "https://" is assumed
//   - Scheme, path, query, fragment, and userinfo are ignored
//   - Port numbers are stripped
//   - Hostnames are lowercased
//   - Trailing dots are removed
//   - Internationalized domain names are normalized to ASCII (punycode)
//
// Examples:
//
//	HostKey("https://Accounts.Google.com/login") == "accounts.google.com"
//	HostKey("google.com:8443")                  == "google.com"
//	HostKey("пример.рф")                        == "xn--e1afmkfd.xn--p1ai"
//
// HostKey returns false if no valid hostname can be extracted.
func HostKey(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}

	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return "", false
	}

	host := strings.ToLower(strings.TrimSuffix(u.Host, "."))

	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if ascii, err := idna.Lookup.ToASCII(host); err == nil && ascii != "" {
		host = ascii
	}

	if host == "" {
		return "", false
	}

	return host, true
}

// SiteKey returns a normalized site identifier for a URL-like string.
//
// For DNS hostnames, the site identifier is the registrable domain
// (effective TLD plus one label) as defined by the Public Suffix List.
// This allows credentials for "google.com" to match
// "accounts.google.com", "mail.google.com", etc.
//
// Special cases:
//   - IP addresses are returned as-is
//   - Single-label hosts (e.g. "localhost") are returned as-is
//   - If the Public Suffix List cannot determine a registrable domain,
//     the normalized host is returned
//
// Examples:
//
//	SiteKey("https://accounts.google.com") == "google.com"
//	SiteKey("google.com")                  == "google.com"
//	SiteKey("foo.github.io")               == "foo.github.io"
//	SiteKey("127.0.0.1:8080")              == "127.0.0.1"
//	SiteKey("localhost")                   == "localhost"
//
// SiteKey returns false if no valid site identifier can be extracted.
func SiteKey(s string) (string, bool) {
	h, ok := HostKey(s)
	if !ok {
		return "", false
	}

	if net.ParseIP(h) != nil {
		return h, true
	}

	if !strings.Contains(h, ".") {
		return h, true
	}

	if rd, err := publicsuffix.EffectiveTLDPlusOne(h); err == nil && rd != "" {
		return rd, true
	}

	return h, true
}
