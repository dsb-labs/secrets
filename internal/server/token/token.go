// Package token provides types for generating and parsing JWT tokens.
package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type (
	// The Token type represents a parsed JWT token.
	Token struct {
		id  uuid.UUID
		raw string
	}
)

// ID returns the subject of the Token.
func (t Token) ID() uuid.UUID {
	return t.id
}

// String returns the raw JWT token string.
func (t Token) String() string {
	return t.raw
}

// Valid returns true if the Token has both an id and raw string representation.
func (t Token) Valid() bool {
	return t.raw != "" && t.id != uuid.Nil
}

type (
	// The Generator type is used to generate JWT tokens.
	Generator struct {
		issuer     string
		ttl        time.Duration
		signingKey []byte
		audience   string
	}

	// The GeneratorConfig type contains fields used to configure a Generator.
	GeneratorConfig struct {
		// The JWT token's issuer.
		Issuer string
		// The TTL of the JWT token.
		TTL time.Duration
		// The key used to sign the JWT token.
		SigningKey []byte
		// The JWT token's audience.
		Audience string
	}
)

// NewGenerator returns a new instance of the Generator type that will generate Token instances using the provided
// configuration.
func NewGenerator(config GeneratorConfig) *Generator {
	return &Generator{
		issuer:     config.Issuer,
		ttl:        config.TTL,
		signingKey: config.SigningKey,
		audience:   config.Audience,
	}
}

// Generate a new Token, using the provided uuid.UUID as the subject.
func (g *Generator) Generate(id uuid.UUID) (Token, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    g.issuer,
		Subject:   id.String(),
		Audience:  []string{g.audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(g.ttl)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.signingKey)
	if err != nil {
		return Token{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return Token{id: id, raw: tokenString}, nil
}

type (
	// The Parser type is responsible for parsing raw Token strings.
	Parser struct {
		signingKey []byte
		parser     *jwt.Parser
	}

	// The ParserConfig type contains fields used to configure a Parser.
	ParserConfig struct {
		// The signing key used when generating the JWT token.
		SigningKey []byte
		// The expected JWT token issuer.
		Issuer string
		// The expected JWT token audience.
		Audience string
	}
)

// NewParser returns a new instance of the Parser type using the provided configuration.
func NewParser(config ParserConfig) *Parser {
	return &Parser{
		signingKey: config.SigningKey,
		parser: jwt.NewParser(
			jwt.WithIssuer(config.Issuer),
			jwt.WithAudience(config.Audience),
			jwt.WithExpirationRequired(),
			jwt.WithLeeway(time.Minute),
		),
	}
}

// Parse the provided token string, returning an instance of the Token type.
func (p *Parser) Parse(tokenString string) (Token, error) {
	var claims jwt.RegisteredClaims
	token, err := p.parser.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return p.signingKey, nil
	})
	switch {
	case err != nil:
		return Token{}, fmt.Errorf("failed to parse token: %w", err)
	case !token.Valid:
		return Token{}, errors.New("invalid token")
	}

	subject, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Token{}, fmt.Errorf("failed to parse token subject: %w", err)
	}

	return Token{id: subject, raw: tokenString}, nil
}

type (
	ctxKey struct{}
)

// FromContext returns a Token from the given context.Context. Callers should call Token.Valid to check if a valid Token
// existed in the context.Context.
func FromContext(ctx context.Context) Token {
	tkn, ok := ctx.Value(ctxKey{}).(Token)
	if !ok {
		return Token{}
	}

	return tkn
}

// ToContext returns a new context.Context containing the provided Token.
func ToContext(ctx context.Context, tkn Token) context.Context {
	return context.WithValue(ctx, ctxKey{}, tkn)
}

// Middleware returns an http.Handler implementation that attempts to parse a Token from a request's Authorization
// header as a bearer token. It does not prevent requests from continuing down to the next handler if a token is
// not present or invalid, this must be handled by the respective HTTP handlers that care about tokens being present.
func Middleware(p *Parser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		bearer := strings.TrimPrefix(header, "Bearer")
		if bearer == "" {
			next.ServeHTTP(w, r)
			return
		}

		tkn, err := p.Parse(strings.TrimSpace(bearer))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if !tkn.Valid() {
			next.ServeHTTP(w, r)
			return
		}

		ctx := ToContext(r.Context(), tkn)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TestToken is a test helper function to create arbitrary Token instances.
func TestToken(t *testing.T, value string) Token {
	t.Helper()

	return Token{id: uuid.New(), raw: value}
}
