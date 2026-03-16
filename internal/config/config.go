package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	DEFAULT_PORT             = "8000"
	DEFAULT_COOKIE_NAME      = "NoterianCookieJWT"
	DEFAULT_COOKIE_TIME_JWT  = 3600
	DEFAULT_SHUTDOWN_TIMEOUT = 5
	DEFAULT_DB_PORT          = "5432"
)

type JWTConfig struct {
	Secret        string
	CookieName    string
	CookieTimeJWT time.Duration
	Secure        bool
}

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Config struct {
	JWT    JWTConfig
	Server ServerConfig
	DB     DBConfig
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET was not found, shutting down...")
	}

	cookieName := os.Getenv("COOKIE_NAME")
	if cookieName == "" {
		cookieName = DEFAULT_COOKIE_NAME
	}

	secure := os.Getenv("IS_SECURE") == "true"

	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	cookieTimeJWT := DEFAULT_COOKIE_TIME_JWT * time.Second
	if strCookieTimeJWT := os.Getenv("COOKIE_TIME_JWT"); strCookieTimeJWT != "" {
		if intCookieTimeJWT, err := strconv.Atoi(strCookieTimeJWT); err != nil {
			cookieTimeJWT = time.Duration(intCookieTimeJWT) * time.Second
		}
	}

	shutdownTimeout := DEFAULT_SHUTDOWN_TIMEOUT * time.Second
	if timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			shutdownTimeout = time.Duration(timeout) * time.Second
		}
	}

	return &Config{
		JWT: JWTConfig{
			Secret:        jwtSecret,
			CookieName:    cookieName,
			CookieTimeJWT: cookieTimeJWT,
			Secure:        secure,
		},
		Server: ServerConfig{
			Port:            port,
			ShutdownTimeout: shutdownTimeout,
		},
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
		},
	}
}
