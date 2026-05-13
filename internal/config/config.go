package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	DEFAULT_JWT_COOKIE_NAME          = "NoterianJWTCookie"
	DEFAULT_JWT_COOKIE_TIME          = 3600
	DEFAULT_SERVER_PORT              = "8000"
	DEFAULT_SERVER_READ_TIMEOUT      = 15
	DEFAULT_SERVER_WRITE_TIMEOUT     = 15
	DEFAULT_SERVER_IDLE_TIMEOUT      = 15
	DEFAULT_SERVER_SHUTDOWN_TIMEOUT  = 5
	DEFAULT_DB_PORT                  = "5432"
	DEFAULT_DB_MAX_OPEN_CONNECTIONS  = 25
	DEFAULT_DB_MAX_IDLE_CONNECTIONS  = 5
	DEFAULT_CSRF_COOKIE_NAME         = "NoterianCSRFCookie"
	DEFAULT_CSRF_COOKIE_TIME         = 3600
	DEFAULT_CSRF_HEADER_NAME         = "X-CSRF-Token"
	DEFAULT_MINIO_ENDPOINT           = "minio:9000"
	DEFAULT_MINIO_USE_SSL            = false
	DEFAULT_MINIO_ATTACHMENTS_BUCKET = "attachments"
	DEFAULT_MINIO_AVATARS_BUCKET     = "avatars"
	DEFAULT_ATTACHMENTS_PORT         = "50051"
	DEFAULT_NOTES_PORT               = "50052"
	DEFAULT_PROFILES_PORT            = "50053"
	DEFAULT_ATTACHMENTS_ADDR         = "localhost:50051"
	DEFAULT_NOTES_ADDR               = "localhost:50052"
	DEFAULT_PROFILES_ADDR            = "localhost:50053"
)

type JWTConfig struct {
	Secret     string
	CookieName string
	CookieTime time.Duration
	Secure     bool
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

type CSRFConfig struct {
	CookieName string
	CookieTime time.Duration
	HeaderName string
	Secure     bool
}

type MinIOConfig struct {
	Endpoint          string
	AccessKey         string
	SecretKey         string
	UseSSL            bool
	AttachmentsBucket string
	AvatarsBucket     string
}

type ServicesConfig struct {
	AttachmentsPort string
	NotesPort       string
	ProfilesPort    string
	AttachmentsAddr string
	NotesAddr       string
	ProfilesAddr    string
}

type Config struct {
	JWT      JWTConfig
	Server   ServerConfig
	DB       DBConfig
	CSRF     CSRFConfig
	MinIO    MinIOConfig
	Services ServicesConfig
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET was not found, shutting down...")
	}

	jwtCookieName := os.Getenv("JWT_COOKIE_NAME")
	if jwtCookieName == "" {
		jwtCookieName = DEFAULT_JWT_COOKIE_NAME
	}

	jwtCookieTime := DEFAULT_JWT_COOKIE_TIME * time.Second
	if strCookieTimeJWT := os.Getenv("JWT_COOKIE_TIME"); strCookieTimeJWT != "" {
		if intCookieTimeJWT, err := strconv.Atoi(strCookieTimeJWT); err == nil {
			jwtCookieTime = time.Duration(intCookieTimeJWT) * time.Second
		}
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = DEFAULT_SERVER_PORT
	}

	serverReadTimeout := DEFAULT_SERVER_READ_TIMEOUT * time.Second
	if strReadTimeout := os.Getenv("SERVER_READ_TIMEOUT"); strReadTimeout != "" {
		if intReadTimeout, err := strconv.Atoi(strReadTimeout); err == nil {
			serverReadTimeout = time.Duration(intReadTimeout) * time.Second
		}
	}

	serverWriteTimeout := DEFAULT_SERVER_WRITE_TIMEOUT * time.Second
	if strWriteTimeout := os.Getenv("SERVER_WRITE_TIMEOUT"); strWriteTimeout != "" {
		if intWriteTimeout, err := strconv.Atoi(strWriteTimeout); err == nil {
			serverWriteTimeout = time.Duration(intWriteTimeout) * time.Second
		}
	}

	serverIdleTimeout := DEFAULT_SERVER_IDLE_TIMEOUT * time.Second
	if strIdleTimeout := os.Getenv("SERVER_IDLE_TIMEOUT"); strIdleTimeout != "" {
		if intIdleTimeout, err := strconv.Atoi(strIdleTimeout); err == nil {
			serverIdleTimeout = time.Duration(intIdleTimeout) * time.Second
		}
	}

	serverShutdownTimeout := DEFAULT_SERVER_SHUTDOWN_TIMEOUT * time.Second
	if timeoutStr := os.Getenv("SERVER_SHUTDOWN_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			serverShutdownTimeout = time.Duration(timeout) * time.Second
		}
	}

	dbMaxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNECTIONS"))
	if err != nil {
		dbMaxOpenConns = DEFAULT_DB_MAX_OPEN_CONNECTIONS
	}

	dbMaxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
	if err != nil {
		dbMaxIdleConns = DEFAULT_DB_MAX_IDLE_CONNECTIONS
	}

	csrfCookieName := os.Getenv("CSRF_COOKIE_NAME")
	if csrfCookieName == "" {
		csrfCookieName = DEFAULT_CSRF_COOKIE_NAME
	}

	csrfCookieTime := DEFAULT_CSRF_COOKIE_TIME * time.Second
	if timeoutStr := os.Getenv("CSRF_COOKIE_TIME"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			csrfCookieTime = time.Duration(timeout) * time.Second
		}
	}

	csrfHeaderName := os.Getenv("CSRF_HEADER_NAME")
	if csrfHeaderName == "" {
		csrfHeaderName = DEFAULT_CSRF_HEADER_NAME
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = DEFAULT_MINIO_ENDPOINT
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	attachmentsBucket := os.Getenv("MINIO_ATTACHMENTS_BUCKET")
	if attachmentsBucket == "" {
		attachmentsBucket = DEFAULT_MINIO_ATTACHMENTS_BUCKET
	}

	avatarsBucket := os.Getenv("MINIO_AVATARS_BUCKET")
	if avatarsBucket == "" {
		avatarsBucket = DEFAULT_MINIO_AVATARS_BUCKET
	}

	attachmentsPort := os.Getenv("ATTACHMENTS_PORT")
	if attachmentsPort == "" {
		attachmentsPort = DEFAULT_ATTACHMENTS_PORT
	}

	notesPort := os.Getenv("NOTES_PORT")
	if notesPort == "" {
		notesPort = DEFAULT_NOTES_PORT
	}

	profilesPort := os.Getenv("PROFILES_PORT")
	if profilesPort == "" {
		profilesPort = DEFAULT_PROFILES_PORT
	}

	attachmentsAddr := os.Getenv("ATTACHMENTS_ADDR")
	if attachmentsAddr == "" {
		attachmentsAddr = DEFAULT_ATTACHMENTS_ADDR
	}

	notesAddr := os.Getenv("NOTES_ADDR")
	if notesAddr == "" {
		notesAddr = DEFAULT_NOTES_ADDR
	}

	profilesAddr := os.Getenv("PROFILES_ADDR")
	if profilesAddr == "" {
		profilesAddr = DEFAULT_PROFILES_ADDR
	}

	secure := os.Getenv("IS_SECURE") == "true"

	return &Config{
		JWT: JWTConfig{
			Secret:     jwtSecret,
			CookieName: jwtCookieName,
			CookieTime: jwtCookieTime,
			Secure:     secure,
		},
		Server: ServerConfig{
			Port:            serverPort,
			ReadTimeout:     serverReadTimeout,
			WriteTimeout:    serverWriteTimeout,
			IdleTimeout:     serverIdleTimeout,
			ShutdownTimeout: serverShutdownTimeout,
		},
		DB: DBConfig{
			Host:         os.Getenv("DB_HOST"),
			Port:         os.Getenv("DB_PORT"),
			User:         os.Getenv("DB_USER"),
			Password:     os.Getenv("DB_PASSWORD"),
			Name:         os.Getenv("DB_NAME"),
			SSLMode:      os.Getenv("DB_SSL_MODE"),
			MaxOpenConns: dbMaxOpenConns,
			MaxIdleConns: dbMaxIdleConns,
		},
		CSRF: CSRFConfig{
			CookieName: csrfCookieName,
			CookieTime: csrfCookieTime,
			HeaderName: csrfHeaderName,
			Secure:     secure,
		},
		MinIO: MinIOConfig{
			Endpoint:          endpoint,
			AccessKey:         accessKey,
			SecretKey:         secretKey,
			UseSSL:            useSSL,
			AttachmentsBucket: attachmentsBucket,
			AvatarsBucket:     avatarsBucket,
		},
		Services: ServicesConfig{
			AttachmentsPort: attachmentsPort,
			NotesPort:       notesPort,
			ProfilesPort:    profilesPort,
			AttachmentsAddr: attachmentsAddr,
			NotesAddr:       notesAddr,
			ProfilesAddr:    profilesAddr,
		},
	}
}
