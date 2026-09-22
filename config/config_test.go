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
	os.Unsetenv("AGRICOLA_URL")
	os.Unsetenv("AGRICOLA_COOKIE")
	os.Unsetenv("TYPESAFE_API_KEY")
	os.Unsetenv("TYPESAFE_BASE_URL")
	os.Unsetenv("TYPESAFE_MODEL")
	os.Unsetenv("TYPESAFE_CLASSIFY_ALL")

	cfg := LoadConfig()

	if cfg.Port != "3000" {
		t.Errorf("expected default Port '3000', got %q", cfg.Port)
	}
	if cfg.RedisURL != "localhost:6379" {
		t.Errorf("expected default RedisURL 'localhost:6379', got %q", cfg.RedisURL)
	}
	if cfg.AgricolaURL != "https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal" {
		t.Errorf("expected default AgricolaURL 'https://www.bancoagricola.com/com/promociones/promociones_get?segmento=principal', got %q", cfg.AgricolaURL)
	}
	if cfg.AgricolaCookie != "" {
		t.Errorf("expected default empty AgricolaCookie, got %q", cfg.AgricolaCookie)
	}
	if cfg.TypeSafeAPIKey != "" {
		t.Errorf("expected default empty TypeSafeAPIKey, got %q", cfg.TypeSafeAPIKey)
	}
	if cfg.TypeSafeBaseURL != "https://api.typesafe.ai/v1/systemone" {
		t.Errorf("expected default TypeSafeBaseURL 'https://api.typesafe.ai/v1/systemone', got %q", cfg.TypeSafeBaseURL)
	}
	if cfg.TypeSafeModel != "jev-latest" {
		t.Errorf("expected default TypeSafeModel 'jev-latest', got %q", cfg.TypeSafeModel)
	}
	if cfg.TypeSafeClassifyAll != false {
		t.Errorf("expected default TypeSafeClassifyAll false, got true")
	}
}

func TestLoadConfigCustomEnv(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("REDIS_URL", "redis.internal:6379")
	t.Setenv("REDIS_PASSWORD", "secret123")
	t.Setenv("CUSCATLAN_API_KEY", "key_abc_123")
	t.Setenv("AGRICOLA_URL", "https://custom.agricola.com/promos")
	t.Setenv("AGRICOLA_COOKIE", "session_cookie=abc")
	t.Setenv("TYPESAFE_API_KEY", "ts_key_123")
	t.Setenv("TYPESAFE_BASE_URL", "https://mock.typesafe.ai/v1/systemone")
	t.Setenv("TYPESAFE_MODEL", "jev-custom")
	t.Setenv("TYPESAFE_CLASSIFY_ALL", "true")

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
	if cfg.AgricolaURL != "https://custom.agricola.com/promos" {
		t.Errorf("expected AgricolaURL 'https://custom.agricola.com/promos', got %q", cfg.AgricolaURL)
	}
	if cfg.AgricolaCookie != "session_cookie=abc" {
		t.Errorf("expected AgricolaCookie 'session_cookie=abc', got %q", cfg.AgricolaCookie)
	}
	if cfg.TypeSafeAPIKey != "ts_key_123" {
		t.Errorf("expected TypeSafeAPIKey 'ts_key_123', got %q", cfg.TypeSafeAPIKey)
	}
	if cfg.TypeSafeBaseURL != "https://mock.typesafe.ai/v1/systemone" {
		t.Errorf("expected TypeSafeBaseURL 'https://mock.typesafe.ai/v1/systemone', got %q", cfg.TypeSafeBaseURL)
	}
	if cfg.TypeSafeModel != "jev-custom" {
		t.Errorf("expected TypeSafeModel 'jev-custom', got %q", cfg.TypeSafeModel)
	}
	if !cfg.TypeSafeClassifyAll {
		t.Errorf("expected TypeSafeClassifyAll true, got false")
	}
}
