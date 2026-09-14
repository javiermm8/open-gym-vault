package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var windowScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return n
`)

type rateLimitConfig struct {
	limit     int
	windowSec int
}

var (
	authLimit    = rateLimitConfig{limit: 5, windowSec: 60}
	defaultLimit = rateLimitConfig{limit: 100, windowSec: 60}
)

var loadEnvLimits = sync.OnceValue(func() struct{} {
	if v := os.Getenv("RATE_LIMIT_AUTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			authLimit.limit = n
		} else {
			log.Printf("rate limiter: ignoring invalid RATE_LIMIT_AUTH=%q", v)
		}
	}
	if v := os.Getenv("RATE_LIMIT_DEFAULT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			defaultLimit.limit = n
		} else {
			log.Printf("rate limiter: ignoring invalid RATE_LIMIT_DEFAULT=%q", v)
		}
	}
	return struct{}{}
})

func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loadEnvLimits()

		cfg := defaultLimit
		if isAuthEndpoint(r.URL.Path) {
			cfg = authLimit
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		now := time.Now().Unix()
		window := now / int64(cfg.windowSec)
		key := "ratelimit:" + scope(r.URL.Path) + ":" + ip + ":" + strconv.FormatInt(window, 10)

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		count, err := windowScript.Run(ctx, s.redisClient, []string{key}, cfg.windowSec*1000+1000).Int()
		if err != nil {
			log.Printf("rate limiter: redis error, allowing request: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		if count > cfg.limit {
			retryAfter := cfg.windowSec - int(now%int64(cfg.windowSec))
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeErrorMessage(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthEndpoint(path string) bool {
	return path == "/auth/login" || path == "/auth/register"
}

func scope(path string) string {
	if isAuthEndpoint(path) {
		return "auth"
	}
	return "default"
}
