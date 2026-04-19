package bot

import (
	"expense-bot/internal/config"
	"testing"

	"gopkg.in/telebot.v4"
)

// mockContext is a simple mock for telebot.Context
type mockContext struct {
	telebot.Context
	sender *telebot.User
}

func (m *mockContext) Sender() *telebot.User {
	return m.sender
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		allowedUsers map[int64]bool
		senderID     int64
		expectNext   bool
	}{
		{
			name:         "Empty allowed users (Fail Closed)",
			allowedUsers: map[int64]bool{},
			senderID:     123,
			expectNext:   false,
		},
		{
			name:         "User is allowed",
			allowedUsers: map[int64]bool{123: true},
			senderID:     123,
			expectNext:   true,
		},
		{
			name:         "User is not allowed",
			allowedUsers: map[int64]bool{456: true},
			senderID:     123,
			expectNext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				AllowedUsers: tt.allowedUsers,
			}

			middleware := AuthMiddleware(cfg)

			var nextCalled bool
			next := func(c telebot.Context) error {
				nextCalled = true
				return nil
			}

			handler := middleware(next)

			ctx := &mockContext{
				sender: &telebot.User{ID: tt.senderID},
			}

			err := handler(ctx)

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if nextCalled != tt.expectNext {
				t.Errorf("expected nextCalled=%v, got %v", tt.expectNext, nextCalled)
			}
		})
	}
}
