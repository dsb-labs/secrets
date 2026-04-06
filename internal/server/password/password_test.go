package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dsb-labs/secrets/internal/server/password"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Options      password.GenerateOptions
		ExpectsError bool
		Check        func(t *testing.T, result string)
	}{
		{
			Name: "generates password with all character sets",
			Options: password.GenerateOptions{
				Length:    16,
				Uppercase: true,
				Lowercase: true,
				Numbers:   true,
				Symbols:   true,
			},
			Check: func(t *testing.T, result string) {
				assert.Len(t, result, 16)
			},
		},
		{
			Name: "respects requested length",
			Options: password.GenerateOptions{
				Length:    32,
				Lowercase: true,
			},
			Check: func(t *testing.T, result string) {
				assert.Len(t, result, 32)
			},
		},
		{
			Name: "generates password with only uppercase",
			Options: password.GenerateOptions{
				Length:    8,
				Uppercase: true,
			},
			Check: func(t *testing.T, result string) {
				assert.Len(t, result, 8)
				assert.Regexp(t, `^[A-Z]+$`, result)
			},
		},
		{
			Name: "generates password with only lowercase",
			Options: password.GenerateOptions{
				Length:    8,
				Lowercase: true,
			},
			Check: func(t *testing.T, result string) {
				assert.Len(t, result, 8)
				assert.Regexp(t, `^[a-z]+$`, result)
			},
		},
		{
			Name: "generates password with only numbers",
			Options: password.GenerateOptions{
				Length:  8,
				Numbers: true,
			},
			Check: func(t *testing.T, result string) {
				assert.Len(t, result, 8)
				assert.Regexp(t, `^[0-9]+$`, result)
			},
		},
		{
			Name: "error if no character sets selected",
			Options: password.GenerateOptions{
				Length: 16,
			},
			ExpectsError: true,
		},
		{
			Name:         "error if length is zero",
			Options:      password.GenerateOptions{Lowercase: true},
			ExpectsError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := password.Generate(tc.Options)
			if tc.ExpectsError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			tc.Check(t, result)
		})
	}
}

func TestRate(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name     string
		Password string
		Expected password.Rating
	}{
		{
			Name:     "very weak password",
			Password: "password",
			Expected: password.RatingVeryWeak,
		},
		{
			Name:     "weak password",
			Password: "Monday99",
			Expected: password.RatingWeak,
		},
		{
			Name:     "good password",
			Password: "Tr0ub4dor",
			Expected: password.RatingGood,
		},
		{
			Name:     "strong password",
			Password: "Tr0ub4dor&",
			Expected: password.RatingStrong,
		},
		{
			Name:     "very strong password",
			Password: "Tr0ub4dor&3",
			Expected: password.RatingVeryStrong,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual := password.Rate(tc.Password)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
