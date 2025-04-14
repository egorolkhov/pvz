package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

//go:generate mockgen -source=tokenManager.go -destination=mocks/tokenManager_mock.go -package=mocks

type TokenManager interface {
	BuildToken(uuid string, role string) (string, error)
	Parse(JwtToken string) (string, string, error)
}

type Manager struct {
	tokenExp  time.Duration
	secretKey string
}

type tokenClaims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManager(secretKey string, tokenExp time.Duration) *Manager {
	return &Manager{
		tokenExp:  tokenExp,
		secretKey: secretKey,
	}
}

func (m *Manager) BuildToken(uuid string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenExp))},
		UserID: uuid,
		Role:   role,
	})
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (m *Manager) Parse(JwtToken string) (string, string, error) {
	token, err := jwt.ParseWithClaims(JwtToken, &tokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return "", "", err
	}
	if !token.Valid {
		return "", "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return "", "", fmt.Errorf("token claims are not of type *tokenClaims")
	}

	return claims.Role, claims.UserID, nil
}
