// Package password provides utilities for evaluating the strength of passwords.
package password

import (
	"github.com/nbutton23/zxcvbn-go"
)

// Rating represents the strength of a password.
type (
	Rating uint8
)

const (
	// RatingVeryWeak indicates a password that is very easily guessed.
	RatingVeryWeak Rating = iota
	// RatingWeak indicates a password that offers minimal protection.
	RatingWeak
	// RatingGood indicates a password with moderate strength.
	RatingGood
	// RatingStrong indicates a password that is difficult to guess.
	RatingStrong
	// RatingVeryStrong indicates a password with the highest level of strength.
	RatingVeryStrong
)

// String returns a human-readable representation of the rating.
func (r Rating) String() string {
	switch r {
	case RatingVeryWeak:
		return "Very Weak"
	case RatingWeak:
		return "Weak"
	case RatingGood:
		return "Good"
	case RatingStrong:
		return "Strong"
	case RatingVeryStrong:
		return "Very Strong"
	default:
		return ""
	}
}

// Rate evaluates the strength of the given password and returns its Rating.
func Rate(password string) Rating {
	strength := zxcvbn.PasswordStrength(password, nil)

	return Rating(strength.Score)
}
