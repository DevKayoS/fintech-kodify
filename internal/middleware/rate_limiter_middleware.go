package middleware

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DevKayoS/fintech-kodify/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err == nil {
			redisClient = redis.NewClient(opt)
			return
		}
	}

	// fallback: variáveis individuais
	redisClient = redis.NewClient(&redis.Options{
		Addr:        os.Getenv("UPSTASH_REDIS_ADDR"),
		Password:    os.Getenv("UPSTASH_REDIS_PASSWORD"),
		TLSConfig:   &tls.Config{},
		DialTimeout: 500 * time.Millisecond,
		MaxRetries:  0,
	})
}

// RateLimiter limita por IP: maxRequests por window.
// Exemplo: RateLimiter(60, time.Minute) → 60 req/min por IP.
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		key := fmt.Sprintf("rl:%s", ip)

		count, err := redisClient.Incr(context.Background(), key).Result()
		if err != nil {
			// se Redis falhar, deixa passar (fail open)
			ctx.Next()
			return
		}

		if count == 1 {
			redisClient.Expire(context.Background(), key, window)
		}

		ttl, _ := redisClient.TTL(context.Background(), key).Result()

		ctx.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, maxRequests-int(count))))
		ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

		if int(count) > maxRequests {
			ctx.Error(errors.ApiError{
				Code:       "TOO_MANY_REQUESTS",
				Message:    "rate limit exceeded, try again later",
				StatusCode: http.StatusTooManyRequests,
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
