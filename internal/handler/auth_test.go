package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"

	"github.com/test-tt/pkg/response"
)

// mockRegister simulates user registration
func mockRegister(c context.Context, ctx *app.RequestContext) {
	var req RegisterRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 1001, "invalid params")
		return
	}

	if req.Email == "existing@example.com" {
		response.Error(ctx, 2005, "email already in use")
		return
	}

	if len(req.Password) < 6 {
		response.Error(ctx, 2009, "password too weak")
		return
	}

	response.Success(ctx, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    1,
			"name":  req.Name,
			"email": req.Email,
		},
		"token": "mock-jwt-token",
	})
}

// mockLogin simulates user login
func mockLogin(c context.Context, ctx *app.RequestContext) {
	var req LoginRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 1001, "invalid params")
		return
	}

	if req.Email == "notfound@example.com" {
		response.Error(ctx, 2001, "user not found")
		return
	}

	if req.Password != "correct123" {
		response.Error(ctx, 2004, "invalid password")
		return
	}

	response.Success(ctx, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    1,
			"name":  "Test User",
			"email": req.Email,
		},
		"token": "mock-jwt-token",
	})
}

// mockLogout simulates user logout
func mockLogout(c context.Context, ctx *app.RequestContext) {
	authHeader := string(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		response.Error(ctx, 1002, "unauthorized")
		return
	}

	response.SuccessWithMessage(ctx, "logged out successfully", nil)
}

// mockGetProfile simulates get user profile
func mockGetProfile(c context.Context, ctx *app.RequestContext) {
	authHeader := string(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		response.Error(ctx, 2008, "login required")
		return
	}

	response.Success(ctx, map[string]interface{}{
		"id":    1,
		"name":  "Test User",
		"email": "test@example.com",
		"age":   25,
	})
}

// mockUpdateProfile simulates update user profile
func mockUpdateProfile(c context.Context, ctx *app.RequestContext) {
	authHeader := string(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		response.Error(ctx, 2008, "login required")
		return
	}

	var req UpdateProfileRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 1001, "invalid params")
		return
	}

	if req.Email == "existing@example.com" {
		response.Error(ctx, 2005, "email already in use")
		return
	}

	response.Success(ctx, map[string]interface{}{
		"id":    1,
		"name":  req.Name,
		"email": req.Email,
		"age":   req.Age,
	})
}

// mockChangePassword simulates password change
func mockChangePassword(c context.Context, ctx *app.RequestContext) {
	authHeader := string(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		response.Error(ctx, 2008, "login required")
		return
	}

	var req ChangePasswordRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 1001, "invalid params")
		return
	}

	if req.OldPassword != "correct123" {
		response.Error(ctx, 2004, "current password is incorrect")
		return
	}

	if len(req.NewPassword) < 6 {
		response.Error(ctx, 2009, "password too weak")
		return
	}

	response.SuccessWithMessage(ctx, "password changed successfully", nil)
}

// mockDeleteAccount simulates account deletion
func mockDeleteAccount(c context.Context, ctx *app.RequestContext) {
	authHeader := string(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		response.Error(ctx, 2008, "login required")
		return
	}

	var req DeleteAccountRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 1001, "invalid params")
		return
	}

	if req.Password != "correct123" {
		response.Error(ctx, 2004, "password is incorrect")
		return
	}

	response.SuccessWithMessage(ctx, "account deleted successfully", nil)
}

// TestRegister tests user registration
func TestRegister(t *testing.T) {
	r := newTestEngine()
	r.POST("/api/v1/auth/register", mockRegister)

	tests := []struct {
		name       string
		body       RegisterRequest
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid registration",
			body:       RegisterRequest{Name: "New User", Email: "new@example.com", Password: "password123"},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "email already exists",
			body:       RegisterRequest{Name: "New User", Email: "existing@example.com", Password: "password123"},
			wantStatus: http.StatusOK,
			wantCode:   2005,
		},
		{
			name:       "password too short",
			body:       RegisterRequest{Name: "New User", Email: "new@example.com", Password: "123"},
			wantStatus: http.StatusOK,
			wantCode:   2009,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodPost, "/api/v1/auth/register",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				ut.Header{Key: "Content-Type", Value: "application/json"},
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestLogin tests user login
func TestLogin(t *testing.T) {
	r := newTestEngine()
	r.POST("/api/v1/auth/login", mockLogin)

	tests := []struct {
		name       string
		body       LoginRequest
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid login",
			body:       LoginRequest{Email: "test@example.com", Password: "correct123"},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "user not found",
			body:       LoginRequest{Email: "notfound@example.com", Password: "password123"},
			wantStatus: http.StatusOK,
			wantCode:   2001,
		},
		{
			name:       "wrong password",
			body:       LoginRequest{Email: "test@example.com", Password: "wrongpassword"},
			wantStatus: http.StatusOK,
			wantCode:   2004,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodPost, "/api/v1/auth/login",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				ut.Header{Key: "Content-Type", Value: "application/json"},
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestLogout tests user logout
func TestLogout(t *testing.T) {
	r := newTestEngine()
	r.POST("/api/v1/auth/logout", mockLogout)

	tests := []struct {
		name       string
		headers    []ut.Header
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid logout",
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "missing authorization",
			headers:    []ut.Header{},
			wantStatus: http.StatusOK,
			wantCode:   1002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ut.PerformRequest(r, http.MethodPost, "/api/v1/auth/logout", nil, tt.headers...)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestGetProfile tests get user profile
func TestGetProfile(t *testing.T) {
	r := newTestEngine()
	r.GET("/api/v1/auth/profile", mockGetProfile)

	tests := []struct {
		name       string
		headers    []ut.Header
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid get profile",
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "missing authorization",
			headers:    []ut.Header{},
			wantStatus: http.StatusOK,
			wantCode:   2008,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ut.PerformRequest(r, http.MethodGet, "/api/v1/auth/profile", nil, tt.headers...)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestUpdateProfile tests update user profile
func TestUpdateProfile(t *testing.T) {
	r := newTestEngine()
	r.PUT("/api/v1/auth/profile", mockUpdateProfile)

	tests := []struct {
		name       string
		body       UpdateProfileRequest
		headers    []ut.Header
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid update",
			body:       UpdateProfileRequest{Name: "Updated Name", Age: 30, Email: "updated@example.com"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "email already exists",
			body:       UpdateProfileRequest{Name: "Updated Name", Email: "existing@example.com"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2005,
		},
		{
			name:       "missing authorization",
			body:       UpdateProfileRequest{Name: "Updated Name"},
			headers:    []ut.Header{{Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2008,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodPut, "/api/v1/auth/profile",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				tt.headers...,
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestChangePassword tests password change
func TestChangePassword(t *testing.T) {
	r := newTestEngine()
	r.PUT("/api/v1/auth/password", mockChangePassword)

	tests := []struct {
		name       string
		body       ChangePasswordRequest
		headers    []ut.Header
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid change password",
			body:       ChangePasswordRequest{OldPassword: "correct123", NewPassword: "newpassword123"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "wrong old password",
			body:       ChangePasswordRequest{OldPassword: "wrongpassword", NewPassword: "newpassword123"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2004,
		},
		{
			name:       "new password too short",
			body:       ChangePasswordRequest{OldPassword: "correct123", NewPassword: "123"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2009,
		},
		{
			name:       "missing authorization",
			body:       ChangePasswordRequest{OldPassword: "correct123", NewPassword: "newpassword123"},
			headers:    []ut.Header{{Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2008,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodPut, "/api/v1/auth/password",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				tt.headers...,
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestDeleteAccount tests account deletion
func TestDeleteAccount(t *testing.T) {
	r := newTestEngine()
	r.DELETE("/api/v1/auth/account", mockDeleteAccount)

	tests := []struct {
		name       string
		body       DeleteAccountRequest
		headers    []ut.Header
		wantStatus int
		wantCode   float64
	}{
		{
			name:       "valid delete account",
			body:       DeleteAccountRequest{Password: "correct123"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   0,
		},
		{
			name:       "wrong password",
			body:       DeleteAccountRequest{Password: "wrongpassword"},
			headers:    []ut.Header{{Key: "Authorization", Value: "Bearer mock-token"}, {Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2004,
		},
		{
			name:       "missing authorization",
			body:       DeleteAccountRequest{Password: "correct123"},
			headers:    []ut.Header{{Key: "Content-Type", Value: "application/json"}},
			wantStatus: http.StatusOK,
			wantCode:   2008,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodDelete, "/api/v1/auth/account",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				tt.headers...,
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// BenchmarkRegister benchmark for registration
func BenchmarkRegister(b *testing.B) {
	r := newTestEngine()
	r.POST("/api/v1/auth/register", mockRegister)

	body, _ := json.Marshal(RegisterRequest{Name: "Test User", Email: "test@example.com", Password: "password123"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ut.PerformRequest(r, http.MethodPost, "/api/v1/auth/register",
			&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"},
		)
	}
}

// BenchmarkLogin benchmark for login
func BenchmarkLogin(b *testing.B) {
	r := newTestEngine()
	r.POST("/api/v1/auth/login", mockLogin)

	body, _ := json.Marshal(LoginRequest{Email: "test@example.com", Password: "correct123"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ut.PerformRequest(r, http.MethodPost, "/api/v1/auth/login",
			&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"},
		)
	}
}
