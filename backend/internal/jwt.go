package internal

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	Authorized bool   `json:"authorized"`
	User       string `json:"user"`
	jwt.StandardClaims
}

func GenerateJWT(username string, key string, expiration time.Duration) (string, error) {
	secretKey := []byte(key)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Authorized: true,
		User:       username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
