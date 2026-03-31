// Package password provides utilities for evaluating the strength of passwords.
package password

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/nbutton23/zxcvbn-go"
)

type (
	// The GenerateOptions type contains fields used to configure password generation.
	GenerateOptions struct {
		// The desired length of the generated password.
		Length int
		// Whether to include uppercase letters.
		Uppercase bool
		// Whether to include lowercase letters.
		Lowercase bool
		// Whether to include numeric digits.
		Numbers bool
		// Whether to include symbols.
		Symbols bool
	}
)

const (
	charsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetLower   = "abcdefghijklmnopqrstuvwxyz"
	charsetNumbers = "0123456789"
	charsetSymbols = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

// Generate creates a cryptographically random password matching the given options. Returns an error if no
// character sets are selected or the requested length is less than one.
func Generate(opts GenerateOptions) (string, error) {
	if opts.Length < 1 {
		return "", errors.New("length must be at least 1")
	}

	charset := ""
	if opts.Uppercase {
		charset += charsetUpper
	}
	if opts.Lowercase {
		charset += charsetLower
	}
	if opts.Numbers {
		charset += charsetNumbers
	}
	if opts.Symbols {
		charset += charsetSymbols
	}

	if charset == "" {
		return "", errors.New("at least one character set must be selected")
	}

	result := make([]byte, opts.Length)
	size := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, size)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}

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
