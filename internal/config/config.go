package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	DEFAULT_PORT                 = "8000"
	DEFAULT_COOKIE_NAME          = "NoterianCookieJWT"
	DEFAULT_COOKIE_TIME_JWT      = 3600
	DEFAULT_SHUTDOWN_TIMEOUT     = 5
	DEFAULT_DB_PORT              = "5432"
	DEFAULT_READ_TIMEOUT         = 15
	DEFAULT_WRITE_TIMEOUT        = 15
	DEFAULT_IDLE_TIMEOUT         = 15
	DEFAULT_MAX_OPEN_CONNECTIONS = 25
	DEFAULT_MAX_IDLE_CONNECTIONS = 5
	DEFAULT_MINIO_ENDPOINT       = "minio:9000"
	DEFAULT_MINIO_USE_SSL        = false
	DEFAULT_MINIO_BUCKET_NAME    = "attachments"
)

type JWTConfig struct {
	Secret        string
	CookieName    string
	CookieTimeJWT time.Duration
	Secure        bool
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
}

type Config struct {
	JWT    JWTConfig
	Server ServerConfig
	DB     DBConfig
	MinIO  MinIOConfig
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
		if intCookieTimeJWT, err := strconv.Atoi(strCookieTimeJWT); err == nil {
			cookieTimeJWT = time.Duration(intCookieTimeJWT) * time.Second
		}
	}

	readTimeout := DEFAULT_READ_TIMEOUT * time.Second
	if strReadTimeout := os.Getenv("READ_TIMEOUT"); strReadTimeout != "" {
		if intReadTimeout, err := strconv.Atoi(strReadTimeout); err == nil {
			readTimeout = time.Duration(intReadTimeout) * time.Second
		}
	}

	writeTimeout := DEFAULT_WRITE_TIMEOUT * time.Second
	if strWriteTimeout := os.Getenv("WRITE_TIMEOUT"); strWriteTimeout != "" {
		if intWriteTimeout, err := strconv.Atoi(strWriteTimeout); err == nil {
			writeTimeout = time.Duration(intWriteTimeout) * time.Second
		}
	}

	idleTimeout := DEFAULT_IDLE_TIMEOUT * time.Second
	if strIdleTimeout := os.Getenv("IDLE_TIMEOUT"); strIdleTimeout != "" {
		if intIdleTimeout, err := strconv.Atoi(strIdleTimeout); err == nil {
			idleTimeout = time.Duration(intIdleTimeout) * time.Second
		}
	}

	shutdownTimeout := DEFAULT_SHUTDOWN_TIMEOUT * time.Second
	if timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			shutdownTimeout = time.Duration(timeout) * time.Second
		}
	}

	maxOpenConns, err := strconv.Atoi(os.Getenv("MAX_OPEN_CONNECTIONS"))
	if err != nil {
		maxOpenConns = DEFAULT_MAX_OPEN_CONNECTIONS
	}

	maxIdleConns, err := strconv.Atoi(os.Getenv("MAX_IDLE_CONNECTIONS"))
	if err != nil {
		maxIdleConns = DEFAULT_MAX_IDLE_CONNECTIONS
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = DEFAULT_MINIO_ENDPOINT
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		bucketName = DEFAULT_MINIO_BUCKET_NAME
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
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		DB: DBConfig{
			Host:         os.Getenv("DB_HOST"),
			Port:         os.Getenv("DB_PORT"),
			User:         os.Getenv("DB_USER"),
			Password:     os.Getenv("DB_PASSWORD"),
			Name:         os.Getenv("DB_NAME"),
			SSLMode:      os.Getenv("DB_SSL_MODE"),
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
		},
		MinIO: MinIOConfig{
			Endpoint:   endpoint,
			AccessKey:  accessKey,
			SecretKey:  secretKey,
			UseSSL:     useSSL,
			BucketName: bucketName,
		},
	}
}
