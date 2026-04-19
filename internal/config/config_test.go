package config

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// The variables read by LoadConfig
	envKeys := []string{
		"ALLOWED_USERS",
		"TELEGRAM_TOKEN",
		"DB_URL",
		"WEBHOOK_URL",
		"WEBHOOK_PORT",
	}

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "all valid fields",
			envVars: map[string]string{
				"TELEGRAM_TOKEN": "mytoken",
				"DB_URL":         "postgres://user:pass@localhost:5432/db",
				"ALLOWED_USERS":  "123,456",
				"WEBHOOK_URL":    "https://example.com",
				"WEBHOOK_PORT":   "8443",
			},
			expected: &Config{
				TelegramToken: "mytoken",
				DBURL:         "postgres://user:pass@localhost:5432/db",
				AllowedUsers:  map[int64]bool{123: true, 456: true},
				WebhookURL:    "https://example.com",
				WebhookPort:   "8443",
			},
		},
		{
			name:    "empty environment",
			envVars: map[string]string{},
			expected: &Config{
				TelegramToken: "",
				DBURL:         "",
				AllowedUsers:  map[int64]bool{},
				WebhookURL:    "",
				WebhookPort:   "",
			},
		},
		{
			name: "allowed users with spaces",
			envVars: map[string]string{
				"ALLOWED_USERS": " 123 ,  456   ",
			},
			expected: &Config{
				TelegramToken: "",
				DBURL:         "",
				AllowedUsers:  map[int64]bool{123: true, 456: true},
				WebhookURL:    "",
				WebhookPort:   "",
			},
		},
		{
			name: "allowed users with invalid data mixed",
			envVars: map[string]string{
				"ALLOWED_USERS": "123, abc, 456, foo, 789",
			},
			expected: &Config{
				TelegramToken: "",
				DBURL:         "",
				AllowedUsers:  map[int64]bool{123: true, 456: true, 789: true},
				WebhookURL:    "",
				WebhookPort:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unset all relevant variables before the test using t.Setenv
			// t.Setenv automatically cleans up after the test.
			for _, key := range envKeys {
				t.Setenv(key, "")
			}

			// Set the variables for the current test
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("LoadConfig() = %v, want %v", cfg, tt.expected)
			}
		})
	}
}
