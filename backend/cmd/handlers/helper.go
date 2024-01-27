package handlers

import (
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/petr-discover/config"
)

func CheckLogin(w http.ResponseWriter, r *http.Request) (string, bool) {
	accessToken, err := r.Cookie("access_token")
	if err != nil || accessToken == nil {
		refreshToken, err := r.Cookie("refresh_token")
		if err == nil && refreshToken != nil {
			username, exists := checkRefreshToken(refreshToken)
			if exists {
				err = handleJWTCookie(w, username)
				if err != nil {
					log.Println("Error generating JWT cookie:", err)
					return "", false
				}
			}
		} else {
			log.Println("Error retrieving refresh token:", err)
			return "", false
		}
	} else {
		username, exists := checkAccessToken(accessToken)
		if !exists {
			log.Println("Error retrieving access token:", err)
			return "", false
		}
		return username, true
	}
	return "", false
}

func checkRefreshToken(tokenVal *http.Cookie) (string, bool) {
	tokenString := tokenVal.Value
	keys := config.JWTSecretKey()
	rKey := keys.RefreshKey
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(rKey), nil
	})
	if err != nil {
		log.Println("Error decoding JWT:", err)
		return "", false
	}
	if tokenVal.Valid() != nil {
		log.Println("Error validating JWT:", err)
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Error retrieving claims")
		return "", false
	}
	log.Println("Claims found in JWT:", claims)
	username, exists := claims["user"].(string)
	if !exists {
		log.Println("User claim not found in JWT")
		return "", false
	}
	return username, exists
}

func checkAccessToken(tokenVal *http.Cookie) (string, bool) {
	tokenString := tokenVal.Value
	keys := config.JWTSecretKey()
	aKey := keys.SecretKey
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(aKey), nil
	})
	if err != nil {
		log.Println("Error decoding JWT:", err)
		return "", false
	}
	if tokenVal.Valid() != nil {
		log.Println("Error validating JWT:", err)
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Error retrieving claims")
		return "", false
	}
	log.Println("Claims found in JWT:", claims)
	username, exists := claims["user"].(string)
	if !exists {
		log.Println("User claim not found in JWT")
		return "", false
	}
	return username, exists
}
