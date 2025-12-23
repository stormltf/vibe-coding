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

// getJWTConfig 返回统一的 JWT 配置
func getJWTConfig() *jwt.Config {
	if config.Cfg != nil && config.Cfg.JWT != nil {
		return &jwt.Config{
			Secret:     config.Cfg.JWT.Secret,
			Issuer:     config.Cfg.JWT.Issuer,
			ExpireTime: config.Cfg.JWT.ExpireTime,
		}
	}
	// 开发环境默认配置
	jwtConfig := jwt.DefaultConfig()
	jwtConfig.Secret = "dev-secret-key-at-least-32-chars!"
	return jwtConfig
}

func Register(h *server.Hertz) {
	// 全局中间件
	// CORS 配置：开发环境允许 localhost，生产环境需要显式配置允许的域名
	var corsConfig *middleware.CORSConfig
	if config.Cfg == nil || config.Cfg.IsDev() {
		corsConfig = middleware.DevCORSConfig()
	} else {
		// 生产环境：从配置加载允许的域名，或使用默认安全配置
		corsConfig = middleware.DefaultCORSConfig()
		// TODO: 从配置文件加载 AllowedOrigins
		// corsConfig.AllowedOrigins = config.Cfg.CORS.AllowedOrigins
	}

	h.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.SecurityHeaders(), // 安全响应头
		middleware.CORSWithConfig(corsConfig),
		gzip.Gzip(gzip.DefaultCompression,
			gzip.WithExcludedExtensions([]string{".png", ".jpg", ".jpeg", ".gif", ".html", ".css", ".js", ".svg"}),
			gzip.WithExcludedPaths([]string{"/"}),
		),
		middleware.Metrics(),
		middleware.AccessLog(),
	)

	pingHandler := handler.NewPingHandler()
	userHandler := handler.NewUserHandler()
	authHandler := handler.NewAuthHandler()
	projectHandler := handler.NewProjectHandler()

	// 静态文件服务 - 手动处理 JS 和 CSS
	h.GET("/static/js/:file", func(ctx context.Context, c *app.RequestContext) {
		file := c.Param("file")
		data, err := os.ReadFile("./web/static/js/" + file)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.Data(http.StatusOK, "application/javascript", data)
	})
	h.GET("/static/css/:file", func(ctx context.Context, c *app.RequestContext) {
		file := c.Param("file")
		data, err := os.ReadFile("./web/static/css/" + file)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Header("Content-Type", "text/css; charset=utf-8")
		c.Data(http.StatusOK, "text/css", data)
	})
	h.GET("/static/assets/:file", func(ctx context.Context, c *app.RequestContext) {
		file := c.Param("file")
		data, err := os.ReadFile("./web/static/assets/" + file)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})
	h.StaticFile("/favicon.ico", "./web/static/assets/favicon.svg")

	// 首页 - 手动设置 Content-Type 避免乱码
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./web/index.html")
	})

	// 工作空间页面
	h.GET("/workspace.html", func(ctx context.Context, c *app.RequestContext) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./web/workspace.html")
	})

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
		// 认证相关 - 公开接口（添加严格限流防止暴力破解）
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthRateLimit()) // 认证端点专用限流：每 IP 每分钟 10 次
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 认证相关 - 需要登录
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.JWTAuth(getJWTConfig()))
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.GET("/profile", authHandler.GetProfile)
			authProtected.PUT("/profile", authHandler.UpdateProfile)
			authProtected.PUT("/password", authHandler.ChangePassword)
			authProtected.DELETE("/account", authHandler.DeleteAccount)
		}

		// 用户相关 - 公开接口
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUserByID)
		}

		// 需要认证的接口
		authUsers := v1.Group("/users")
		authUsers.Use(middleware.JWTAuth(getJWTConfig()))
		{
			authUsers.POST("", userHandler.CreateUser)
			authUsers.PUT("/:id", userHandler.UpdateUser)
			authUsers.DELETE("/:id", userHandler.DeleteUser)
		}

		// 项目相关 - 需要认证
		projects := v1.Group("/projects")
		projects.Use(middleware.JWTAuth(getJWTConfig()))
		{
			projects.GET("", projectHandler.List)
			projects.POST("", projectHandler.Create)
			projects.GET("/:id", projectHandler.Get)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
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
	c             *app.RequestContext
	header        http.Header
	headerWritten bool
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

// syncHeaders 同步所有 headers 到 Hertz（只执行一次）
func (r *responseWriterAdapter) syncHeaders() {
	if r.headerWritten {
		return
	}
	r.headerWritten = true

	// 同步所有 header 值（支持多值 header）
	for k, values := range r.header {
		for i, v := range values {
			if i == 0 {
				r.c.Response.Header.Set(k, v)
			} else {
				r.c.Response.Header.Add(k, v)
			}
		}
	}
}

func (r *responseWriterAdapter) Write(data []byte) (int, error) {
	r.syncHeaders()
	return r.c.Write(data)
}

func (r *responseWriterAdapter) WriteHeader(statusCode int) {
	r.syncHeaders()
	r.c.SetStatusCode(statusCode)
}

// Flush 实现 http.Flusher 接口
func (r *responseWriterAdapter) Flush() {
	r.syncHeaders()
	// Hertz 会自动处理 flush
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
// 安全要求：仅支持 Authorization Header，不支持 Query 参数（避免 Token 泄露到日志）
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

		// 仅支持 Authorization header（安全考虑：Query 参数会被记录到访问日志）
		auth := string(c.GetHeader("Authorization"))
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    4001,
				"message": "unauthorized: Authorization header required",
			})
			return
		}

		// Bearer token 格式
		if len(auth) > 7 && auth[:7] == "Bearer " {
			auth = auth[7:]
		}

		// 使用常量时间比较防止时序攻击
		if !secureCompare(auth, token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    4001,
				"message": "unauthorized: invalid debug token",
			})
			return
		}

		c.Next(ctx)
	}
}

// secureCompare 常量时间字符串比较，防止时序攻击
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
