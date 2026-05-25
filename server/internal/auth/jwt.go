package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, role, secret string) (string, error) {
	if role == "" {
		role = "user"
	}
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (string, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(*jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := t.Claims.(*Claims); ok && t.Valid {
		return claims.UserID, nil
	}
	return "", jwt.ErrSignatureInvalid
}
