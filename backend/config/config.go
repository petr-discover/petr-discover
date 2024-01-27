package config

import (
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var Neo4jUser string
var Neo4jPass string

type AppConfig struct {
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
	DBType string
}

type JWTConfig struct {
	SecretKey  string
	RefreshKey string
}

var GoogleURLAPI string

var GoogleOauthConfig oauth2.Config

func GoogleAuthConfig() {
	loadEnv()
	GoogleOauthConfig = oauth2.Config{
		RedirectURL:  "http://localhost:8080/api/v1/auth/google/callback",
		ClientID:     getEnv("CLIENT_ID", "some_id"),
		ClientSecret: getEnv("CLIENT_SECRET", "some_secret"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	GoogleURLAPI = getEnv("GOOGLE_URL_API", "https://www.googleapis.com/oauth2/v2/userinfo?access_token=")
}

func Neo4jDBConfig() string {
	loadEnv()
	cfg := &AppConfig{
		DBHost: getEnv("NDB_HOST", "localhost"),
		DBPort: getEnvInt("NDB_PORT", 7687),
		DBUser: getEnv("NDB_USER", "neo4j"),
		DBPass: getEnv("NDB_PASS", "123456789a"),
		DBType: getEnv("NDB_TYPE", "bolt"),
	}
	Neo4jUser = cfg.DBUser
	Neo4jPass = cfg.DBPass
	uri := fmt.Sprintf("%s://%s:%d", cfg.DBType, cfg.DBHost, cfg.DBPort)
	return uri
}

func PostgresDBConfig() string {
	loadEnv()
	cfg := &AppConfig{
		DBHost: getEnv("PDB_HOST", "localhost"),
		DBPort: getEnvInt("PDB_PORT", 5432),
		DBUser: getEnv("PDB_USER", "irvine"),
		DBPass: getEnv("PDB_PASS", "irvine"),
		DBName: getEnv("PDB_NAME", "hackathon"),
		DBType: getEnv("PDB_TYPE", "postgres"),
	}
	dbURL := fmt.Sprintf("%s://%s:%s@%s:%d/%s", cfg.DBType, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	return dbURL
}

func JWTSecretKey() *JWTConfig {
	loadEnv()
	cfg := &JWTConfig{
		SecretKey:  getEnv("JWT_SECRET_KEY", "secret"),
		RefreshKey: getEnv("JWT_REFReSH_KEY", "refresh"),
	}
	return cfg
}
