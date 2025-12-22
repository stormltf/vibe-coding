package validate

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// Init 初始化验证器
func Init() {
	once.Do(func() {
		validate = validator.New()

		// 使用 json tag 作为字段名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
}

// Struct 验证结构体
func Struct(s interface{}) error {
	if validate == nil {
		Init()
	}
	return validate.Struct(s)
}

// Var 验证单个变量
func Var(field interface{}, tag string) error {
	if validate == nil {
		Init()
	}
	return validate.Var(field, tag)
}

// ValidationErrors 将验证错误转换为友好的错误信息
func ValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errs[e.Field()] = getErrorMsg(e)
		}
	}
	return errs
}

// FirstError 获取第一个错误信息
func FirstError(err error) string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrs) > 0 {
			return getErrorMsg(validationErrs[0])
		}
	}
	return err.Error()
}

func getErrorMsg(e validator.FieldError) string {
	field := e.Field()
	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + e.Param()
	case "max":
		return field + " must be at most " + e.Param()
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "oneof":
		return field + " must be one of: " + e.Param()
	default:
		return field + " is invalid"
	}
}
