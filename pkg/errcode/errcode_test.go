package errcode

import (
	"net/http"
	"testing"
)

func TestErrCode_Error(t *testing.T) {
	err := ErrNotFound
	if err.Error() != "not found" {
		t.Errorf("Error() = %s, want %s", err.Error(), "not found")
	}
}

func TestErrCode_WithMessage(t *testing.T) {
	err := ErrNotFound.WithMessage("user not found")

	if err.Code != ErrNotFound.Code {
		t.Errorf("Code = %d, want %d", err.Code, ErrNotFound.Code)
	}

	if err.Message != "user not found" {
		t.Errorf("Message = %s, want %s", err.Message, "user not found")
	}

	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusNotFound)
	}
}

func TestPreDefinedErrors(t *testing.T) {
	tests := []struct {
		err        *ErrCode
		wantCode   int
		wantStatus int
	}{
		{Success, 0, http.StatusOK},
		{ErrInvalidParams, 1001, http.StatusBadRequest},
		{ErrUnauthorized, 1002, http.StatusUnauthorized},
		{ErrNotFound, 1004, http.StatusNotFound},
		{ErrUserNotFound, 2001, http.StatusNotFound},
	}

	for _, tt := range tests {
		if tt.err.Code != tt.wantCode {
			t.Errorf("%v.Code = %d, want %d", tt.err, tt.err.Code, tt.wantCode)
		}
		if tt.err.HTTPStatus != tt.wantStatus {
			t.Errorf("%v.HTTPStatus = %d, want %d", tt.err, tt.err.HTTPStatus, tt.wantStatus)
		}
	}
}
