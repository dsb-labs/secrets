package password_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidsbond/keeper/internal/server/password"
)

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
