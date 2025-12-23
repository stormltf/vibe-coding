package service

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestPasswordHashing tests password hashing and verification
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Test hashing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Verify the hash is not the same as the password
	if string(hashedPassword) == password {
		t.Error("hashed password should not equal original password")
	}

	// Test verification with correct password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Errorf("password verification failed for correct password: %v", err)
	}

	// Test verification with wrong password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Error("password verification should fail for wrong password")
	}
}

// TestPasswordLength tests minimum password length validation
func TestPasswordLength(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"12345", false},
		{"123456", true},
		{"password123", true},
		{"verylongpasswordthatismorethan128characters" +
			"verylongpasswordthatismorethan128characters" +
			"verylongpasswordthatismorethan128characters", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			valid := len(tt.password) >= minPasswordLength
			if valid != tt.valid {
				t.Errorf("password %q: got valid=%v, want valid=%v", tt.password, valid, tt.valid)
			}
		})
	}
}

// TestTokenBlacklistKey tests token blacklist key format
func TestTokenBlacklistKey(t *testing.T) {
	tests := []struct {
		token string
		want  string
	}{
		{"abc123", "token:blacklist:abc123"},
		{"jwt-token-here", "token:blacklist:jwt-token-here"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			got := testTokenBlacklistKey(tt.token)
			if got != tt.want {
				t.Errorf("tokenBlacklistKey(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func testTokenBlacklistKey(token string) string {
	return "token:blacklist:" + token
}

// TestErrorMessages tests error message constants
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		err     error
		wantMsg string
	}{
		{ErrUserNotFound, "user not found"},
		{ErrInvalidPassword, "invalid password"},
		{ErrEmailExists, "email already exists"},
		{ErrPasswordTooShort, "password too short"},
		{ErrTokenBlacklisted, "token is blacklisted"},
	}

	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

// BenchmarkPasswordHashing benchmarks password hashing
func BenchmarkPasswordHashing(b *testing.B) {
	password := []byte("testpassword123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	}
}

// BenchmarkPasswordVerification benchmarks password verification
func BenchmarkPasswordVerification(b *testing.B) {
	password := []byte("testpassword123")
	hashedPassword, _ := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcrypt.CompareHashAndPassword(hashedPassword, password)
	}
}
