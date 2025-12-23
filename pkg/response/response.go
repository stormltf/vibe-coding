package response

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"github.com/test-tt/pkg/errcode"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *app.RequestContext, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *app.RequestContext, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

func Error(c *app.RequestContext, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func ErrorWithStatus(c *app.RequestContext, statusCode int, code int, message string) {
	c.JSON(statusCode, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Fail 使用 ErrCode 返回错误
func Fail(c *app.RequestContext, err *errcode.ErrCode) {
	c.JSON(err.HTTPStatus, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}

// FailWithData 使用 ErrCode 返回错误，附带数据
func FailWithData(c *app.RequestContext, err *errcode.ErrCode, data interface{}) {
	c.JSON(err.HTTPStatus, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	})
}
