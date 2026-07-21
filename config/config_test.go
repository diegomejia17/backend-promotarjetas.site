package config

import (
	"os"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("CUSCATLAN_API_KEY")

	cfg := LoadConfig()

	if cfg.Port != "3000" {
		t.Errorf("expected default Port '3000', got %q", cfg.Port)
	}
	if cfg.RedisURL != "localhost:6379" {
		t.Errorf("expected default RedisURL 'localhost:6379', got %q", cfg.RedisURL)
	}
}

func TestLoadConfigCustomEnv(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("REDIS_URL", "redis.internal:6379")
	t.Setenv("REDIS_PASSWORD", "secret123")
	t.Setenv("CUSCATLAN_API_KEY", "key_abc_123")

	cfg := LoadConfig()

	if cfg.Port != "8080" {
		t.Errorf("expected Port '8080', got %q", cfg.Port)
	}
	if cfg.RedisURL != "redis.internal:6379" {
		t.Errorf("expected RedisURL 'redis.internal:6379', got %q", cfg.RedisURL)
	}
	if cfg.RedisPassword != "secret123" {
		t.Errorf("expected RedisPassword 'secret123', got %q", cfg.RedisPassword)
	}
	if cfg.CuscatlanAPIKey != "key_abc_123" {
		t.Errorf("expected CuscatlanAPIKey 'key_abc_123', got %q", cfg.CuscatlanAPIKey)
	}
}
