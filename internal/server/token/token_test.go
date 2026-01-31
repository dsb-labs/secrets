package token_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidsbond/keeper/internal/server/token"
)

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generates a valid token", func(t *testing.T) {
		generator := token.NewGenerator(token.GeneratorConfig{
			Issuer:     "test.com",
			TTL:        time.Hour,
			SigningKey: bytes.Repeat([]byte{0}, 32),
			Audience:   "test.com",
		})

		tkn, err := generator.Generate(uuid.NameSpaceDNS)
		require.NoError(t, err)
		assert.EqualValues(t, uuid.NameSpaceDNS, tkn.ID())
		assert.NotEmpty(t, tkn.String())
		assert.True(t, tkn.Valid())
	})
}

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	const (
		issuer   = "test.com"
		audience = "test.com"
	)

	key := bytes.Repeat([]byte{0}, 32)

	generator := token.NewGenerator(token.GeneratorConfig{
		Issuer:     issuer,
		TTL:        time.Hour,
		SigningKey: key,
		Audience:   audience,
	})

	parser := token.NewParser(token.ParserConfig{
		SigningKey: key,
		Issuer:     issuer,
		Audience:   audience,
	})

	t.Run("parses a valid token", func(t *testing.T) {
		expected, err := generator.Generate(uuid.NameSpaceDNS)
		require.NoError(t, err)

		actual, err := parser.Parse(expected.String())
		require.NoError(t, err)

		assert.EqualValues(t, expected.String(), actual.String())
		assert.EqualValues(t, expected.ID(), actual.ID())
		assert.True(t, actual.Valid())
	})

	t.Run("error for invalid token", func(t *testing.T) {
		actual, err := parser.Parse("invalid")
		require.Error(t, err)
		assert.False(t, actual.Valid())
	})
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name         string
		Token        func(g *token.Generator) string
		ExpectsToken bool
		ExpectedID   uuid.UUID
	}{
		{
			Name: "ignore no token header",
		},
		{
			Name: "ignore invalid token in header",
			Token: func(g *token.Generator) string {
				return "invalid"
			},
		},
		{
			Name: "propagate valid token",
			Token: func(g *token.Generator) string {
				tkn, err := g.Generate(uuid.NameSpaceDNS)
				require.NoError(t, err)
				return tkn.String()
			},
			ExpectsToken: true,
			ExpectedID:   uuid.NameSpaceDNS,
		},
	}

	for _, tc := range tt {
		parser := token.NewParser(token.ParserConfig{
			SigningKey: bytes.Repeat([]byte{0}, 32),
			Issuer:     "test.com",
			Audience:   "test.com",
		})

		generator := token.NewGenerator(token.GeneratorConfig{
			Issuer:     "test.com",
			TTL:        time.Hour,
			SigningKey: bytes.Repeat([]byte{0}, 32),
			Audience:   "test.com",
		})

		handler := token.Middleware(parser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tkn := token.FromContext(r.Context())
			if tc.ExpectsToken {
				assert.Equal(t, tc.ExpectedID, tkn.ID())
				assert.True(t, tkn.Valid())
				return
			}

			assert.False(t, tkn.Valid())
		}))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		if tc.Token != nil {
			if tkn := tc.Token(generator); tkn != "" {
				r.Header.Set("Authorization", "Bearer "+tkn)
			}
		}

		handler.ServeHTTP(w, r)
	}
}
