package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

var (
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
)

var (
	// HTTPRequestTotal 记录 HTTP 请求总数
	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration 记录请求持续时间
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// RPCRequestTotal 记录 RPC 请求总数
	RPCRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_requests_total",
			Help: "Total number of RPC requests",
		},
		[]string{"method", "service"},
	)

	// RPCRequestDuration 记录 RPC 请求持续时间
	RPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rpc_request_duration_seconds",
			Help:    "RPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "service"},
	)

	// 业务指标 - 用户相关
	UserRegistered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of registered users",
		},
	)

	UserLoginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_total",
			Help: "Total number of user login attempts",
		},
		[]string{"status"}, // success, failed
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of currently active users",
		},
	)

	// 业务指标 - 帖子相关
	PostCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "posts_created_total",
			Help: "Total number of created posts",
		},
	)

	PostsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "posts_total",
			Help: "Total number of posts",
		},
	)

	PostOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "post_operations_total",
			Help: "Total number of post operations",
		},
		[]string{"operation"}, // create, update, delete, like
	)

	// 系统指标
	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Number of active database connections",
		},
		[]string{"database"}, // mysql, redis
	)

	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache"}, // redis
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache"}, // redis
	)

	// 错误指标
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"service", "type"}, // service: api-gateway, user-service, post-service; type: db, cache, network, etc
	)
)

// MetricsMiddleware 指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// 增加正在处理的请求数
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// 处理请求
		c.Next()

		// 记录请求持续时间
		duration := time.Since(start).Seconds()
		HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		// 记录请求总数
		HTTPRequestTotal.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()
	}
}

// PrometheusMiddleware 用于收集 HTTP 请求指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录请求持续时间
		duration := time.Since(start).Seconds()

		// 记录请求总数和状态码
		HTTPRequestTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			string(rune(c.Writer.Status())),
		).Inc()

		// 记录请求持续时间
		HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		// 记录错误
		if c.Writer.Status() >= 400 {
			ErrorsTotal.WithLabelValues(
				"api-gateway",
				"http",
			).Inc()
		}
	}
}

// MetricsHandler 返回 metrics 处理函数
func MetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// DatabaseMetricsMiddleware 用于收集数据库指标
func DatabaseMetricsMiddleware(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("metrics:create", func(db *gorm.DB) {
		if db.Statement.Table == "posts" {
			PostOperations.WithLabelValues("create").Inc()
			PostsTotal.Inc()
		}
	})

	db.Callback().Delete().After("gorm:delete").Register("metrics:delete", func(db *gorm.DB) {
		if db.Statement.Table == "posts" {
			PostOperations.WithLabelValues("delete").Inc()
			PostsTotal.Dec()
		}
	})
}

// CacheMetricsMiddleware 用于收集缓存指标
func CacheMetricsMiddleware(key string, hit bool) {
	if hit {
		CacheHits.WithLabelValues("redis").Inc()
	} else {
		CacheMisses.WithLabelValues("redis").Inc()
	}
}

// UserActivityMiddleware 用于记录用户活动
func UserActivityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 如果是登录请求
		if c.FullPath() == "/api/v1/user/login" {
			if c.Writer.Status() == 200 {
				UserLoginTotal.WithLabelValues("success").Inc()
			} else {
				UserLoginTotal.WithLabelValues("failed").Inc()
			}
		}

		// 如果是注册请求
		if c.FullPath() == "/api/v1/user/register" && c.Writer.Status() == 200 {
			UserRegistered.Inc()
		}
	}
}
