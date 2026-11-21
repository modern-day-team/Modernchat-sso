package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Limiter 限流判定接口（fail-open：内部故障时实现方应放行）。
type Limiter interface {
	Allow(key string) bool
}

// LimiterStore 内存令牌桶限流：每 key 一个桶，每 10 分钟清理一次闲置桶。
// 多实例共享配额需换 Redis 实现（主仓库有 middleware/limiter_redis.go 可参照）。
type LimiterStore struct {
	mu      sync.Mutex
	buckets map[string]*rate.Limiter
	perMin  rate.Limit
	burst   int
	lastGC  time.Time
}

// NewLimiterStore 构造：perMin 每分钟次数，burst 桶容量。
func NewLimiterStore(perMin, burst int) *LimiterStore {
	return &LimiterStore{
		buckets: make(map[string]*rate.Limiter),
		perMin:  rate.Limit(perMin) / 60.0,
		burst:   burst,
		lastGC:  time.Now(),
	}
}

// Allow 判定是否放行。
func (l *LimiterStore) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Since(l.lastGC) > 10*time.Minute {
		l.buckets = make(map[string]*rate.Limiter)
		l.lastGC = time.Now()
	}
	b, ok := l.buckets[key]
	if !ok {
		b = rate.NewLimiter(l.perMin, l.burst)
		l.buckets[key] = b
	}
	return b.Allow()
}

// RateLimit 限流中间件：超限 429（文案与主仓库一致）。
func RateLimit(l Limiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(keyFn(c)) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "操作过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}

// KeyByIP 按 IP 限流（认证类公开端点）。
func KeyByIP(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// KeyByUID 按用户限流（登录态端点；未登录回退 IP）。
func KeyByUID(c *gin.Context) string {
	if u := UserFrom(c); u != nil {
		return "u:" + strconv.FormatUint(uint64(u.ID), 10)
	}
	return "ip:" + c.ClientIP()
}
