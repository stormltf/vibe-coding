package errcode

import "net/http"

type ErrCode struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *ErrCode) Error() string {
	return e.Message
}

// 定义错误码
var (
	// 成功
	Success = &ErrCode{Code: 0, Message: "success", HTTPStatus: http.StatusOK}

	// 通用错误 1xxx
	ErrInvalidParams   = &ErrCode{Code: 1001, Message: "invalid params", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized    = &ErrCode{Code: 1002, Message: "unauthorized", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden       = &ErrCode{Code: 1003, Message: "forbidden", HTTPStatus: http.StatusForbidden}
	ErrNotFound        = &ErrCode{Code: 1004, Message: "not found", HTTPStatus: http.StatusNotFound}
	ErrInternalServer  = &ErrCode{Code: 1005, Message: "internal server error", HTTPStatus: http.StatusInternalServerError}
	ErrTooManyRequests = &ErrCode{Code: 1006, Message: "too many requests", HTTPStatus: http.StatusTooManyRequests}

	// 用户相关 2xxx
	ErrUserNotFound      = &ErrCode{Code: 2001, Message: "user not found", HTTPStatus: http.StatusNotFound}
	ErrUserAlreadyExists = &ErrCode{Code: 2002, Message: "user already exists", HTTPStatus: http.StatusConflict}
	ErrInvalidUserID     = &ErrCode{Code: 2003, Message: "invalid user id", HTTPStatus: http.StatusBadRequest}

	// 数据库相关 3xxx
	ErrDatabase = &ErrCode{Code: 3001, Message: "database error", HTTPStatus: http.StatusInternalServerError}

	// 缓存相关 4xxx
	ErrCache = &ErrCode{Code: 4001, Message: "cache error", HTTPStatus: http.StatusInternalServerError}
)

// WithMessage 返回带自定义消息的错误码
func (e *ErrCode) WithMessage(msg string) *ErrCode {
	return &ErrCode{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
	}
}
