// Package card provides views for listing and inspecting stored payment card records.
package card

// cardExpiryLabel returns the expiry label for a list item, using "Expired" when the card has
// expired and "Expires" otherwise.
func cardExpiryLabel(item Item) string {
	if item.Expired {
		return "Expired " + item.Expiry
	}
	return "Expires " + item.Expiry
}

// issuerIconURL returns the local asset URL for the payment icon that corresponds to the given
// issuer short name (as returned by github.com/durango/go-credit-card). Returns an empty
// string when no icon is available for the issuer.
func issuerIconURL(issuer string) string {
	icons := map[string]string{
		"visa":                      "visa",
		"visa electron":             "visa",
		"mastercard":                "mastercard",
		"amex":                      "amex",
		"discover":                  "discover",
		"maestro":                   "maestro",
		"jcb":                       "jcb",
		"diners club international": "diners",
		"diners club carte blanche": "diners",
		"diners club enroute":       "diners",
		"china unionpay":            "unionpay",
		"elo":                       "elo",
		"hipercard":                 "hipercard",
	}

	if name, ok := icons[issuer]; ok {
		return "/asset/image/" + name + ".svg"
	}

	return ""
}
