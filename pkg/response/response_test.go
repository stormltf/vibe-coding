package response

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/test-tt/pkg/errcode"
)

func newTestEngine() *route.Engine {
	opt := config.NewOptions([]config.Option{})
	return route.NewEngine(opt)
}

func TestSuccess(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		Success(ctx, map[string]string{"key": "value"})
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, 0, resp.Code)
	assert.DeepEqual(t, "success", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestSuccessWithMessage(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		SuccessWithMessage(ctx, "custom message", nil)
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, 0, resp.Code)
	assert.DeepEqual(t, "custom message", resp.Message)
}

func TestError(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		Error(ctx, 4001, "validation error")
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, 4001, resp.Code)
	assert.DeepEqual(t, "validation error", resp.Message)
}

func TestErrorWithStatus(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		ErrorWithStatus(ctx, http.StatusNotFound, 4004, "not found")
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, 4004, resp.Code)
	assert.DeepEqual(t, "not found", resp.Message)
}

func TestFail(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		Fail(ctx, errcode.ErrInvalidParams)
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, errcode.ErrInvalidParams.HTTPStatus, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, errcode.ErrInvalidParams.Code, resp.Code)
	assert.DeepEqual(t, errcode.ErrInvalidParams.Message, resp.Message)
}

func TestFailWithData(t *testing.T) {
	r := newTestEngine()
	r.GET("/test", func(c context.Context, ctx *app.RequestContext) {
		FailWithData(ctx, errcode.ErrInvalidParams, map[string]string{"field": "name"})
	})

	w := ut.PerformRequest(r, http.MethodGet, "/test", nil)

	assert.DeepEqual(t, errcode.ErrInvalidParams.HTTPStatus, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Nil(t, err)
	assert.DeepEqual(t, errcode.ErrInvalidParams.Code, resp.Code)
	assert.NotNil(t, resp.Data)
}
