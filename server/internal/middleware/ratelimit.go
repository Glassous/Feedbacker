package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	buckets   map[string]*rateBucket
	lastSweep time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:     limit,
		window:    window,
		buckets:   make(map[string]*rateBucket),
		lastSweep: time.Now(),
	}
}

func (l *rateLimiter) allow(key string) (bool, int, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > 10*l.window {
		for k, b := range l.buckets {
			if now.After(b.resetAt) {
				delete(l.buckets, k)
			}
		}
		l.lastSweep = now
	}

	bucket, ok := l.buckets[key]
	if !ok || now.After(bucket.resetAt) {
		bucket = &rateBucket{count: 0, resetAt: now.Add(l.window)}
		l.buckets[key] = bucket
	}

	if bucket.count >= l.limit {
		return false, 0, time.Until(bucket.resetAt)
	}

	bucket.count++
	return true, l.limit - bucket.count, 0
}

// RateLimit 基于客户端 IP 的固定窗口限流。
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)

	return func(c *gin.Context) {
		allowed, remaining, retryAfter := limiter.allow(c.ClientIP())
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			seconds := int(retryAfter.Seconds())
			if seconds < 1 {
				seconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(seconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    "rate_limited",
				"message": "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
