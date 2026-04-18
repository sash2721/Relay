package middlewares

import (
	"net/http"
	"sync"
	"time"
)

var (
	rateLimitMap = make(map[string]time.Time)
	rateLimitMu  sync.Mutex
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userID").(string)

		rateLimitMu.Lock()
		lastDeploymentTime, ok := rateLimitMap[userId]

		if ok && time.Since(lastDeploymentTime) < 2*time.Minute {
			rateLimitMu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"code":429,"message":"Too many requests. Please wait 2 minutes between deployments."}`))
			return
		}

		rateLimitMap[userId] = time.Now()
		rateLimitMu.Unlock()

		next.ServeHTTP(w, r)
	})
}
