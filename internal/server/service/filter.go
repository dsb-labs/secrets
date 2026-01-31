package service

import (
	"slices"
	"strings"

	"github.com/davidsbond/x/filter"

	"github.com/davidsbond/keeper/internal/server/urlcmp"
)

// LoginsByDomain returns a filter.Filter implementation that checks if a given Login contains a domain that matches
// the one specified. Domains are compared by generating stable host/site keys which allows for flexibility such as
// accounts.google.com matching a domain of google.com.
func LoginsByDomain(domain string) filter.Filter[Login] {
	want, ok := urlcmp.SiteKey(domain)

	return func(login Login) bool {
		if strings.TrimSpace(domain) == "" {
			return true
		}

		if !ok {
			return false
		}

		return slices.ContainsFunc(login.Domains, func(s string) bool {
			have, ok := urlcmp.SiteKey(s)
			return ok && have == want
		})
	}
}

// NotesByQuery returns a filter.Filter implementation that filters notes based on a given query value. The filter
// returns true if either the name or content of the note contains the query text. This filter does not match on
// casing.
func NotesByQuery(query string) filter.Filter[Note] {
	return func(note Note) bool {
		return strings.Contains(strings.ToLower(note.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(note.Content), strings.ToLower(query))
	}
}
