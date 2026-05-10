package middleware

import (
	"net/http"
	"sync"
	"time"

	"suberes_golang/i18n"

	"github.com/gin-gonic/gin"
)

type rateLimiterEntry struct {
	count    int
	expireAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rateLimiterEntry
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rateLimiterEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.visitors {
			if now.After(entry.expireAt) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.visitors[ip]

	if !exists || now.After(entry.expireAt) {
		rl.visitors[ip] = &rateLimiterEntry{
			count:    1,
			expireAt: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.isAllowed(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgTooManyRequests),
				"status":         "failure",
			})
			return
		}
		c.Next()
	}
}
