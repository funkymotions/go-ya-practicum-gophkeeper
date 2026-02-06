package utils

import (
	"github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

func IssueJWTToken(userID int, secret []byte) ([]byte, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyClaims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserID:           userID,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, err
	}

	return []byte(tokenString), nil
}

func CheckJWTToken(tokenString string, secret []byte) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(t *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
