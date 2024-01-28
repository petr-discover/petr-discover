package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/petr-discover/cmd/database"
	"github.com/petr-discover/config"
	"github.com/petr-discover/internal"
	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	Password string `json:"sub"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type RegistRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func AuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var registrationRequest RegistRequest
	err := json.NewDecoder(r.Body).Decode(&registrationRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if registrationRequest.Email == "" || registrationRequest.Username == "" {
		http.Error(w, "Both email and username needs to exist", http.StatusBadRequest)
		return
	}

	userInfo := UserInfo(registrationRequest)
	err = storeUserInDatabase(userInfo)
	if err != nil {
		http.Error(w, "Failed to store user in the database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("message : success"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	var isValidUser bool
	var username string

	if loginRequest.Username != "" {
		isValidUser = authenticateUserByUsername(loginRequest.Username, loginRequest.Password)
		username = loginRequest.Username
	} else if loginRequest.Email != "" {
		username, isValidUser = authenticateUserByEmail(loginRequest.Email, loginRequest.Password)
	} else {
		http.Error(w, "Invalid request. Provide either username or email.", http.StatusBadRequest)
		return
	}

	if isValidUser {
		err = handleJWTCookie(w, username)
		if err != nil {
			http.Error(w, "Error generating JWT cookie", http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("message : success"))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("message : Invalid credentials"))
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("message : Success"))
}

func GoogleAuth(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	config.GoogleAuthConfig()
	url := config.GoogleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	oauthstate, err := r.Cookie("oauthstate")
	if err != nil {
		log.Printf("Error retrieving oauthstate cookie: %v\n", err)
		w.Write([]byte(err.Error()))
		return
	}

	if r.FormValue("state") != oauthstate.Value {
		log.Printf("Invalid google oauth state cookie: %s state: %s\n", oauthstate.Value, r.FormValue("state"))
		w.Write([]byte(err.Error()))
		return
	}

	data, err := getGoogleUserInfo(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	userInfo, err := parseUserInfo(data)
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	username := removeGmailSuffix(userInfo.Email)
	userInfo.Username = username
	err = storeUserInDatabase(userInfo)
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	err = handleJWTCookie(w, userInfo.Username)
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("message : Success"))
	}

}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, cookie)
	return state
}

func getGoogleUserInfo(code string) ([]byte, error) {
	token, err := config.GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange: %s", err.Error())
	}

	fmt.Println(config.GoogleURLAPI + token.AccessToken)
	resp, err := http.Get(config.GoogleURLAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get userInfo: %s", err.Error())
	}

	defer resp.Body.Close()
	fmt.Println(resp.Body, resp.Status)
	return io.ReadAll(resp.Body)
}

func parseUserInfo(data []byte) (UserInfo, error) {
	var userInfo UserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return UserInfo{}, fmt.Errorf("failed to parse user information: %s", err.Error())
	}
	return userInfo, nil
}

func storeUserInDatabase(userInfo UserInfo) error {
	existingUser, err := getUserByUsername(userInfo.Username)
	if err != nil {
		return fmt.Errorf("failed to retrieve user: %s", err.Error())
	}

	if existingUser == nil {
		hashedPassword, err := HashPassword(userInfo.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %s", err.Error())
		}
		_, err = database.DBMain.Exec("INSERT INTO member (username, password, email) VALUES ($1, $2, $3)", userInfo.Username, hashedPassword, userInfo.Email)
		if err != nil {
			return fmt.Errorf("failed to insert new user: %s", err.Error())
		}

		fmt.Println("User created successfully")
		return nil
	} else {
		return fmt.Errorf("user already exists")
	}
}

func authenticateUserByUsername(username, password string) bool {
	query := "SELECT password FROM member WHERE username = $1"
	var hashedPassword string
	err := database.DBMain.QueryRow(query, username).Scan(&hashedPassword)
	if err != nil {
		log.Println(err)
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println(err)
		return false
	}
	return err == nil
}

func authenticateUserByEmail(email, password string) (string, bool) {
	query := "SELECT password, username FROM member WHERE email = $1"
	var hashedPassword string
	var username string
	err := database.DBMain.QueryRow(query, email).Scan(&hashedPassword, &username)
	if err != nil {
		log.Println(err)
		return "", false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println(err)
		return "", false
	}
	return username, err == nil
}

func handleJWTCookie(w http.ResponseWriter, username string) error {
	accessToken, err := internal.GenerateJWT(username, config.JWTSecretKey().SecretKey, 15*time.Minute)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return err
	}

	refreshToken, err := internal.GenerateJWT(username, config.JWTSecretKey().RefreshKey, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
	return nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %s", err.Error())
	}
	return string(hashedPassword), nil
}

func removeGmailSuffix(email string) string {
	username := strings.Split(email, "@")
	temp := username[0]
	return temp
}

func getUserByUsername(username string) (*UserInfo, error) {
	var user UserInfo
	err := database.DBMain.QueryRow("SELECT email, password FROM member WHERE username = $1", username).
		Scan(&user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("database error: %s", err.Error())
	}
	return &user, nil
}
