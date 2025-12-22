package router

import (
	"context"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	"github.com/test-tt/config"
	"github.com/test-tt/internal/handler"
	"github.com/test-tt/internal/middleware"
	"github.com/test-tt/pkg/jwt"
)

func Register(h *server.Hertz) {
	// 全局中间件
	h.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.CORS(),
		gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".png", ".jpg", ".jpeg", ".gif"})),
		middleware.Metrics(),
		middleware.AccessLog(),
	)

	pingHandler := handler.NewPingHandler()
	userHandler := handler.NewUserHandler()

	// 健康检查
	h.GET("/ping", pingHandler.Ping)
	h.GET("/health", pingHandler.Health) // 详细健康检查

	// Prometheus 指标（生产环境建议添加认证）
	if config.Cfg != nil && config.Cfg.IsProd() {
		h.GET("/metrics", debugAuthMiddleware(), prometheusHandler())
	} else {
		h.GET("/metrics", prometheusHandler())
	}

	// Swagger API 文档（仅开发环境）
	if config.Cfg == nil || config.Cfg.IsDev() {
		h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))
	}

	// pprof 性能分析（开发环境直接访问，生产环境需要认证）
	pprofGroup := h.Group("/debug/pprof")
	if config.Cfg != nil && config.Cfg.IsProd() {
		pprofGroup.Use(debugAuthMiddleware())
	}
	{
		pprofGroup.GET("/", pprofHandler(pprof.Index))
		pprofGroup.GET("/cmdline", pprofHandler(pprof.Cmdline))
		pprofGroup.GET("/profile", pprofHandler(pprof.Profile))
		pprofGroup.GET("/symbol", pprofHandler(pprof.Symbol))
		pprofGroup.GET("/trace", pprofHandler(pprof.Trace))
		pprofGroup.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
		pprofGroup.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
		pprofGroup.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
		pprofGroup.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
		pprofGroup.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
		pprofGroup.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	}

	// API v1 - 公开接口
	v1 := h.Group("/api/v1")
	{
		// 用户相关 - 公开接口
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUserByID)
		}

		// 需要认证的接口
		authUsers := v1.Group("/users")
		authUsers.Use(middleware.JWTAuth(jwt.DefaultConfig()))
		{
			authUsers.POST("", userHandler.CreateUser)
			authUsers.PUT("/:id", userHandler.UpdateUser)
			authUsers.DELETE("/:id", userHandler.DeleteUser)
		}
	}
}

// 对象池复用，减少 GC 压力
var (
	headerPool = sync.Pool{
		New: func() interface{} {
			return make(http.Header, 8)
		},
	}
	requestPool = sync.Pool{
		New: func() interface{} {
			return &http.Request{}
		},
	}
)

// prometheusHandler 将 promhttp.Handler 适配为 Hertz handler
func prometheusHandler() app.HandlerFunc {
	h := promhttp.Handler()
	return func(ctx context.Context, c *app.RequestContext) {
		// 从池中获取对象
		req := requestPool.Get().(*http.Request)
		header := headerPool.Get().(http.Header)

		// 清空并复用 header
		for k := range header {
			delete(header, k)
		}

		// 填充请求
		req.Method = string(c.Method())
		req.RequestURI = string(c.URI().RequestURI())
		req.Header = header

		h.ServeHTTP(newResponseWriterAdapter(c), req)

		// 归还到池
		requestPool.Put(req)
		headerPool.Put(header)
	}
}

// responseWriterAdapter 适配 Hertz 的 ResponseWriter
type responseWriterAdapter struct {
	c      *app.RequestContext
	header http.Header
}

func newResponseWriterAdapter(c *app.RequestContext) *responseWriterAdapter {
	return &responseWriterAdapter{
		c:      c,
		header: make(http.Header),
	}
}

func (r *responseWriterAdapter) Header() http.Header {
	return r.header
}

func (r *responseWriterAdapter) Write(data []byte) (int, error) {
	// 在首次写入前同步 headers 到 Hertz
	for k, v := range r.header {
		if len(v) > 0 {
			r.c.Response.Header.Set(k, v[0])
		}
	}
	return r.c.Write(data)
}

func (r *responseWriterAdapter) WriteHeader(statusCode int) {
	// 同步 headers 到 Hertz
	for k, v := range r.header {
		if len(v) > 0 {
			r.c.Response.Header.Set(k, v[0])
		}
	}
	r.c.SetStatusCode(statusCode)
}

// pprofHandler 将 pprof handler 适配为 Hertz handler
func pprofHandler(h http.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 构造 *url.URL
		u := &url.URL{
			Scheme:   string(c.URI().Scheme()),
			Host:     string(c.Host()),
			Path:     string(c.URI().Path()),
			RawQuery: string(c.URI().QueryString()),
		}
		h.ServeHTTP(newResponseWriterAdapter(c), &http.Request{
			Method:     string(c.Method()),
			RequestURI: string(c.URI().RequestURI()),
			URL:        u,
		})
	}
}

// debugAuthMiddleware 调试端点认证中间件
// 用于保护 pprof 和 metrics 等敏感端点
// 通过环境变量 DEBUG_AUTH_TOKEN 设置访问令牌
func debugAuthMiddleware() app.HandlerFunc {
	token := os.Getenv("DEBUG_AUTH_TOKEN")
	tokenRequired := token != "" // 如果设置了 token 则必须验证

	return func(ctx context.Context, c *app.RequestContext) {
		// 如果没有配置 token，生产环境拒绝访问
		if !tokenRequired {
			c.AbortWithStatusJSON(http.StatusForbidden, map[string]interface{}{
				"code":    4003,
				"message": "debug endpoints disabled: DEBUG_AUTH_TOKEN not configured",
			})
			return
		}

		// 检查 Authorization header
		auth := string(c.GetHeader("Authorization"))
		if auth == "" {
			// 也支持 query parameter
			auth = c.Query("token")
		}

		// Bearer token 格式
		if len(auth) > 7 && auth[:7] == "Bearer " {
			auth = auth[7:]
		}

		if auth != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    4001,
				"message": "unauthorized: invalid or missing debug token",
			})
			return
		}

		c.Next(ctx)
	}
}
