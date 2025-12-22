package router

import (
	"context"
	"net/http"
	"net/http/pprof"
	"net/url"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/gzip"
	"github.com/hertz-contrib/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
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

	// Prometheus 指标
	h.GET("/metrics", prometheusHandler())

	// Swagger API 文档
	h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))

	// pprof 性能分析（仅开发/调试环境启用）
	pprofGroup := h.Group("/debug/pprof")
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
