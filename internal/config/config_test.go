package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	origEnv := map[string]string{
		"JWT_SECRET":           os.Getenv("JWT_SECRET"),
		"COOKIE_NAME":          os.Getenv("COOKIE_NAME"),
		"IS_SECURE":            os.Getenv("IS_SECURE"),
		"PORT":                 os.Getenv("PORT"),
		"COOKIE_TIME_JWT":      os.Getenv("COOKIE_TIME_JWT"),
		"READ_TIMEOUT":         os.Getenv("READ_TIMEOUT"),
		"WRITE_TIMEOUT":        os.Getenv("WRITE_TIMEOUT"),
		"IDLE_TIMEOUT":         os.Getenv("IDLE_TIMEOUT"),
		"SHUTDOWN_TIMEOUT":     os.Getenv("SHUTDOWN_TIMEOUT"),
		"MAX_OPEN_CONNECTIONS": os.Getenv("MAX_OPEN_CONNECTIONS"),
		"MAX_IDLE_CONNECTIONS": os.Getenv("MAX_IDLE_CONNECTIONS"),
		"DB_HOST":              os.Getenv("DB_HOST"),
		"DB_PORT":              os.Getenv("DB_PORT"),
		"DB_USER":              os.Getenv("DB_USER"),
		"DB_PASSWORD":          os.Getenv("DB_PASSWORD"),
		"DB_NAME":              os.Getenv("DB_NAME"),
		"DB_SSL_MODE":          os.Getenv("DB_SSL_MODE"),
	}

	defer func() {
		for k, v := range origEnv {
			os.Setenv(k, v)
		}
	}()

	t.Run("uses default values when env vars not set", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("JWT_SECRET", "test-secret-key")

		cfg := Load()

		if cfg.JWT.Secret != "test-secret-key" {
			t.Errorf("expected secret test-secret, got %s", cfg.JWT.Secret)
		}
		if cfg.JWT.CookieName != DEFAULT_COOKIE_NAME {
			t.Errorf("expected cookie name %s, got %s", DEFAULT_COOKIE_NAME, cfg.JWT.CookieName)
		}
		if cfg.JWT.CookieTimeJWT != DEFAULT_COOKIE_TIME_JWT*time.Second {
			t.Errorf("expected cookie time %v, got %v", DEFAULT_COOKIE_TIME_JWT*time.Second, cfg.JWT.CookieTimeJWT)
		}
		if cfg.JWT.Secure != false {
			t.Errorf("expected secure false, got %v", cfg.JWT.Secure)
		}
		if cfg.Server.Port != DEFAULT_PORT {
			t.Errorf("expected port %s, got %s", DEFAULT_PORT, cfg.Server.Port)
		}
		if cfg.Server.ReadTimeout != DEFAULT_READ_TIMEOUT*time.Second {
			t.Errorf("expected read timeout %v, got %v", DEFAULT_READ_TIMEOUT*time.Second, cfg.Server.ReadTimeout)
		}
		if cfg.Server.WriteTimeout != DEFAULT_WRITE_TIMEOUT*time.Second {
			t.Errorf("expected write timeout %v, got %v", DEFAULT_WRITE_TIMEOUT*time.Second, cfg.Server.WriteTimeout)
		}
		if cfg.Server.IdleTimeout != DEFAULT_IDLE_TIMEOUT*time.Second {
			t.Errorf("expected idle timeout %v, got %v", DEFAULT_IDLE_TIMEOUT*time.Second, cfg.Server.IdleTimeout)
		}
		if cfg.Server.ShutdownTimeout != DEFAULT_SHUTDOWN_TIMEOUT*time.Second {
			t.Errorf("expected shutdown timeout %v, got %v", DEFAULT_SHUTDOWN_TIMEOUT*time.Second, cfg.Server.ShutdownTimeout)
		}

		if cfg.DB.MaxOpenConns != DEFAULT_MAX_OPEN_CONNECTIONS {
			t.Errorf("expected max open conns %d, got %d", DEFAULT_MAX_OPEN_CONNECTIONS, cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != DEFAULT_MAX_IDLE_CONNECTIONS {
			t.Errorf("expected max idle conns %d, got %d", DEFAULT_MAX_IDLE_CONNECTIONS, cfg.DB.MaxIdleConns)
		}
	})

	t.Run("uses custom values from env vars", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("JWT_SECRET", "custom-secret-key")
		os.Setenv("COOKIE_NAME", "custom-cookie")
		os.Setenv("IS_SECURE", "true")
		os.Setenv("PORT", "9000")
		os.Setenv("COOKIE_TIME_JWT", "7200")
		os.Setenv("READ_TIMEOUT", "30")
		os.Setenv("WRITE_TIMEOUT", "30")
		os.Setenv("IDLE_TIMEOUT", "30")
		os.Setenv("SHUTDOWN_TIMEOUT", "10")
		os.Setenv("MAX_OPEN_CONNECTIONS", "50")
		os.Setenv("MAX_IDLE_CONNECTIONS", "10")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "postgres")
		os.Setenv("DB_PASSWORD", "password")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSL_MODE", "disable")

		cfg := Load()

		if cfg.JWT.Secret != "custom-secret-key" {
			t.Errorf("expected secret custom-secret, got %s", cfg.JWT.Secret)
		}
		if cfg.JWT.CookieName != "custom-cookie" {
			t.Errorf("expected cookie name custom-cookie, got %s", cfg.JWT.CookieName)
		}
		if cfg.JWT.CookieTimeJWT != 7200*time.Second {
			t.Errorf("expected cookie time 7200s, got %v", cfg.JWT.CookieTimeJWT)
		}
		if cfg.JWT.Secure != true {
			t.Errorf("expected secure true, got %v", cfg.JWT.Secure)
		}

		if cfg.Server.Port != "9000" {
			t.Errorf("expected port 9000, got %s", cfg.Server.Port)
		}
		if cfg.Server.ReadTimeout != 30*time.Second {
			t.Errorf("expected read timeout 30s, got %v", cfg.Server.ReadTimeout)
		}
		if cfg.Server.WriteTimeout != 30*time.Second {
			t.Errorf("expected write timeout 30s, got %v", cfg.Server.WriteTimeout)
		}
		if cfg.Server.IdleTimeout != 30*time.Second {
			t.Errorf("expected idle timeout 30s, got %v", cfg.Server.IdleTimeout)
		}
		if cfg.Server.ShutdownTimeout != 10*time.Second {
			t.Errorf("expected shutdown timeout 10s, got %v", cfg.Server.ShutdownTimeout)
		}

		if cfg.DB.Host != "localhost" {
			t.Errorf("expected host localhost, got %s", cfg.DB.Host)
		}
		if cfg.DB.Port != "5433" {
			t.Errorf("expected port 5433, got %s", cfg.DB.Port)
		}
		if cfg.DB.User != "postgres" {
			t.Errorf("expected user postgres, got %s", cfg.DB.User)
		}
		if cfg.DB.Password != "password" {
			t.Errorf("expected password password, got %s", cfg.DB.Password)
		}
		if cfg.DB.Name != "testdb" {
			t.Errorf("expected name testdb, got %s", cfg.DB.Name)
		}
		if cfg.DB.SSLMode != "disable" {
			t.Errorf("expected sslmode disable, got %s", cfg.DB.SSLMode)
		}
		if cfg.DB.MaxOpenConns != 50 {
			t.Errorf("expected max open conns 50, got %d", cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != 10 {
			t.Errorf("expected max idle conns 10, got %d", cfg.DB.MaxIdleConns)
		}
	})

	t.Run("handles invalid numeric values - uses defaults", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("COOKIE_TIME_JWT", "invalid")
		os.Setenv("READ_TIMEOUT", "invalid")
		os.Setenv("WRITE_TIMEOUT", "invalid")
		os.Setenv("IDLE_TIMEOUT", "invalid")
		os.Setenv("SHUTDOWN_TIMEOUT", "invalid")
		os.Setenv("MAX_OPEN_CONNECTIONS", "invalid")
		os.Setenv("MAX_IDLE_CONNECTIONS", "invalid")

		cfg := Load()

		if cfg.JWT.CookieTimeJWT != DEFAULT_COOKIE_TIME_JWT*time.Second {
			t.Errorf("expected default cookie time %v, got %v", DEFAULT_COOKIE_TIME_JWT*time.Second, cfg.JWT.CookieTimeJWT)
		}
		if cfg.Server.ReadTimeout != DEFAULT_READ_TIMEOUT*time.Second {
			t.Errorf("expected default read timeout %v, got %v", DEFAULT_READ_TIMEOUT*time.Second, cfg.Server.ReadTimeout)
		}
		if cfg.Server.WriteTimeout != DEFAULT_WRITE_TIMEOUT*time.Second {
			t.Errorf("expected default write timeout %v, got %v", DEFAULT_WRITE_TIMEOUT*time.Second, cfg.Server.WriteTimeout)
		}
		if cfg.Server.IdleTimeout != DEFAULT_IDLE_TIMEOUT*time.Second {
			t.Errorf("expected default idle timeout %v, got %v", DEFAULT_IDLE_TIMEOUT*time.Second, cfg.Server.IdleTimeout)
		}
		if cfg.Server.ShutdownTimeout != DEFAULT_SHUTDOWN_TIMEOUT*time.Second {
			t.Errorf("expected default shutdown timeout %v, got %v", DEFAULT_SHUTDOWN_TIMEOUT*time.Second, cfg.Server.ShutdownTimeout)
		}
		if cfg.DB.MaxOpenConns != DEFAULT_MAX_OPEN_CONNECTIONS {
			t.Errorf("expected default max open conns %d, got %d", DEFAULT_MAX_OPEN_CONNECTIONS, cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != DEFAULT_MAX_IDLE_CONNECTIONS {
			t.Errorf("expected default max idle conns %d, got %d", DEFAULT_MAX_IDLE_CONNECTIONS, cfg.DB.MaxIdleConns)
		}
	})
}
