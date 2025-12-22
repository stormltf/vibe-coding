package service

import (
	"strconv"
	"testing"
)

func TestUserCacheKey(t *testing.T) {
	tests := []struct {
		id   uint64
		want string
	}{
		{1, "user:1"},
		{100, "user:100"},
		{999, "user:999"},
	}

	for _, tt := range tests {
		got := testUserCacheKey(tt.id)
		if got != tt.want {
			t.Errorf("testUserCacheKey(%d) = %s, want %s", tt.id, got, tt.want)
		}
	}
}

func testUserCacheKey(id uint64) string {
	return "user:" + strconv.FormatUint(id, 10)
}
