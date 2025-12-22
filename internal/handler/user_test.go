package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/test-tt/pkg/response"
)

// newTestEngine 创建测试用的路由引擎
func newTestEngine() *route.Engine {
	opt := config.NewOptions([]config.Option{})
	opt.Addr = ":8888"
	return route.NewEngine(opt)
}

// mockGetUsers 模拟获取用户列表
func mockGetUsers(c context.Context, ctx *app.RequestContext) {
	response.Success(ctx, map[string]interface{}{
		"list":      []interface{}{},
		"total":     0,
		"page":      1,
		"page_size": 10,
	})
}

// mockGetUserByID 模拟获取用户详情
func mockGetUserByID(c context.Context, ctx *app.RequestContext) {
	idStr := ctx.Param("id")
	if idStr == "" || idStr == "abc" {
		response.Error(ctx, 4001, "invalid id")
		return
	}
	response.Success(ctx, map[string]interface{}{
		"id":    1,
		"name":  "测试用户",
		"age":   25,
		"email": "test@example.com",
	})
}

// mockCreateUser 模拟创建用户
func mockCreateUser(c context.Context, ctx *app.RequestContext) {
	var req CreateUserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 4001, err.Error())
		return
	}
	response.Success(ctx, map[string]interface{}{
		"id":    1,
		"name":  req.Name,
		"age":   req.Age,
		"email": req.Email,
	})
}

// mockUpdateUser 模拟更新用户
func mockUpdateUser(c context.Context, ctx *app.RequestContext) {
	response.Success(ctx, nil)
}

// mockDeleteUser 模拟删除用户
func mockDeleteUser(c context.Context, ctx *app.RequestContext) {
	response.Success(ctx, nil)
}

// TestGetUsers 测试获取用户列表
func TestGetUsers(t *testing.T) {
	r := newTestEngine()
	r.GET("/api/v1/users", mockGetUsers)

	w := ut.PerformRequest(r, http.MethodGet, "/api/v1/users?page=1&page_size=10", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, float64(0), resp["code"])
}

// TestGetUserByID 测试获取单个用户
func TestGetUserByID(t *testing.T) {
	r := newTestEngine()
	r.GET("/api/v1/users/:id", mockGetUserByID)

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   float64
	}{
		{"valid id", "1", http.StatusOK, 0},
		{"invalid id", "abc", http.StatusOK, 4001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ut.PerformRequest(r, http.MethodGet, "/api/v1/users/"+tt.id, nil)
			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, tt.wantCode, resp["code"])
		})
	}
}

// TestCreateUser 测试创建用户
func TestCreateUser(t *testing.T) {
	r := newTestEngine()
	r.POST("/api/v1/users", mockCreateUser)

	tests := []struct {
		name       string
		body       CreateUserRequest
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       CreateUserRequest{Name: "测试用户", Age: 25, Email: "test@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "minimal request",
			body:       CreateUserRequest{Name: "张三"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			w := ut.PerformRequest(r, http.MethodPost, "/api/v1/users",
				&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
				ut.Header{Key: "Content-Type", Value: "application/json"},
			)

			assert.DeepEqual(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.DeepEqual(t, float64(0), resp["code"])
		})
	}
}

// TestUpdateUser 测试更新用户
func TestUpdateUser(t *testing.T) {
	r := newTestEngine()
	r.PUT("/api/v1/users/:id", mockUpdateUser)

	body, _ := json.Marshal(UpdateUserRequest{Name: "更新名称", Age: 30})
	w := ut.PerformRequest(r, http.MethodPut, "/api/v1/users/1",
		&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"},
	)

	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// TestDeleteUser 测试删除用户
func TestDeleteUser(t *testing.T) {
	r := newTestEngine()
	r.DELETE("/api/v1/users/:id", mockDeleteUser)

	w := ut.PerformRequest(r, http.MethodDelete, "/api/v1/users/1", nil)
	assert.DeepEqual(t, http.StatusOK, w.Code)
}

// BenchmarkGetUsers 基准测试 - 获取用户列表
func BenchmarkGetUsers(b *testing.B) {
	r := newTestEngine()
	r.GET("/api/v1/users", mockGetUsers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ut.PerformRequest(r, http.MethodGet, "/api/v1/users?page=1&page_size=10", nil)
	}
}

// BenchmarkCreateUser 基准测试 - 创建用户
func BenchmarkCreateUser(b *testing.B) {
	r := newTestEngine()
	r.POST("/api/v1/users", mockCreateUser)

	body, _ := json.Marshal(CreateUserRequest{Name: "测试用户", Age: 25, Email: "test@example.com"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ut.PerformRequest(r, http.MethodPost, "/api/v1/users",
			&ut.Body{Body: bytes.NewReader(body), Len: len(body)},
			ut.Header{Key: "Content-Type", Value: "application/json"},
		)
	}
}
