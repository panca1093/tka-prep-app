package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yourorg/tkaprep/apps/backend/internal/domain"
)

type contextKey string

const contextKeyClaims contextKey = "jwt_claims"

type Claims struct {
	UserID uuid.UUID
	Role   domain.Role
}

type jwtClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

func GenerateAccessToken(secret string, ttl time.Duration, user *domain.User) (string, error) {
	now := time.Now()
	c := jwtClaims{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, c)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken returns a cryptographically random 32-byte hex string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func ParseAccessToken(secret, tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token expired: %w", err)
		}
		return nil, fmt.Errorf("parse token: %w", err)
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	id, err := uuid.Parse(c.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	return &Claims{
		UserID: id,
		Role:   domain.Role(c.Role),
	}, nil
}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKeyClaims, claims)
}

func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(contextKeyClaims).(*Claims)
	return c, ok
}
