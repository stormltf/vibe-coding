package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/test-tt/pkg/i18n"
)

// I18n 国际化中间件
// 从请求头 Accept-Language 解析语言并注入到 context
func I18n() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 优先从 query 参数获取
		lang := c.Query("lang")

		// 其次从请求头获取
		if lang == "" {
			acceptLang := string(c.GetHeader("Accept-Language"))
			lang = i18n.ParseAcceptLanguage(acceptLang)
		}

		// 注入到 context
		ctx = i18n.WithLang(ctx, lang)
		c.Set("lang", lang)

		c.Next(ctx)
	}
}

// GetLang 从 RequestContext 获取语言
func GetLang(c *app.RequestContext) string {
	if lang, exists := c.Get("lang"); exists {
		return lang.(string)
	}
	return i18n.DefaultLang
}
