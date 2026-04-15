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
		if cfg.JWT.CookieName != DEFAULT_JWT_COOKIE_NAME {
			t.Errorf("expected cookie name %s, got %s", DEFAULT_JWT_COOKIE_NAME, cfg.JWT.CookieName)
		}
		if cfg.JWT.CookieTime != DEFAULT_JWT_COOKIE_TIME*time.Second {
			t.Errorf("expected cookie time %v, got %v", DEFAULT_JWT_COOKIE_TIME*time.Second, cfg.JWT.CookieTime)
		}
		if cfg.JWT.Secure != false {
			t.Errorf("expected secure false, got %v", cfg.JWT.Secure)
		}
		if cfg.Server.Port != DEFAULT_SERVER_PORT {
			t.Errorf("expected port %s, got %s", DEFAULT_SERVER_PORT, cfg.Server.Port)
		}
		if cfg.Server.ReadTimeout != DEFAULT_SERVER_READ_TIMEOUT*time.Second {
			t.Errorf("expected read timeout %v, got %v", DEFAULT_SERVER_READ_TIMEOUT*time.Second, cfg.Server.ReadTimeout)
		}
		if cfg.Server.WriteTimeout != DEFAULT_SERVER_WRITE_TIMEOUT*time.Second {
			t.Errorf("expected write timeout %v, got %v", DEFAULT_SERVER_WRITE_TIMEOUT*time.Second, cfg.Server.WriteTimeout)
		}
		if cfg.Server.IdleTimeout != DEFAULT_SERVER_IDLE_TIMEOUT*time.Second {
			t.Errorf("expected idle timeout %v, got %v", DEFAULT_SERVER_IDLE_TIMEOUT*time.Second, cfg.Server.IdleTimeout)
		}
		if cfg.Server.ShutdownTimeout != DEFAULT_SERVER_SHUTDOWN_TIMEOUT*time.Second {
			t.Errorf("expected shutdown timeout %v, got %v", DEFAULT_SERVER_SHUTDOWN_TIMEOUT*time.Second, cfg.Server.ShutdownTimeout)
		}

		if cfg.DB.MaxOpenConns != DEFAULT_DB_MAX_OPEN_CONNECTIONS {
			t.Errorf("expected max open conns %d, got %d", DEFAULT_DB_MAX_OPEN_CONNECTIONS, cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != DEFAULT_DB_MAX_IDLE_CONNECTIONS {
			t.Errorf("expected max idle conns %d, got %d", DEFAULT_DB_MAX_IDLE_CONNECTIONS, cfg.DB.MaxIdleConns)
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

		if cfg.JWT.CookieTime != DEFAULT_JWT_COOKIE_TIME*time.Second {
			t.Errorf("expected default cookie time %v, got %v", DEFAULT_JWT_COOKIE_TIME*time.Second, cfg.JWT.CookieTime)
		}
		if cfg.Server.ReadTimeout != DEFAULT_SERVER_READ_TIMEOUT*time.Second {
			t.Errorf("expected default read timeout %v, got %v", DEFAULT_SERVER_READ_TIMEOUT*time.Second, cfg.Server.ReadTimeout)
		}
		if cfg.Server.WriteTimeout != DEFAULT_SERVER_WRITE_TIMEOUT*time.Second {
			t.Errorf("expected default write timeout %v, got %v", DEFAULT_SERVER_WRITE_TIMEOUT*time.Second, cfg.Server.WriteTimeout)
		}
		if cfg.Server.IdleTimeout != DEFAULT_SERVER_IDLE_TIMEOUT*time.Second {
			t.Errorf("expected default idle timeout %v, got %v", DEFAULT_SERVER_IDLE_TIMEOUT*time.Second, cfg.Server.IdleTimeout)
		}
		if cfg.Server.ShutdownTimeout != DEFAULT_SERVER_SHUTDOWN_TIMEOUT*time.Second {
			t.Errorf("expected default shutdown timeout %v, got %v", DEFAULT_SERVER_SHUTDOWN_TIMEOUT*time.Second, cfg.Server.ShutdownTimeout)
		}
		if cfg.DB.MaxOpenConns != DEFAULT_DB_MAX_OPEN_CONNECTIONS {
			t.Errorf("expected default max open conns %d, got %d", DEFAULT_DB_MAX_OPEN_CONNECTIONS, cfg.DB.MaxOpenConns)
		}
		if cfg.DB.MaxIdleConns != DEFAULT_DB_MAX_IDLE_CONNECTIONS {
			t.Errorf("expected default max idle conns %d, got %d", DEFAULT_DB_MAX_IDLE_CONNECTIONS, cfg.DB.MaxIdleConns)
		}
	})
}
