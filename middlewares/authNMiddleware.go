package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/sash2721/Relay/errors"
)

func AuthNMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authzToken := r.Header.Get("Authorization")

		// Also check query param for SSE connections
		if authzToken == "" {
			if qToken := r.URL.Query().Get("token"); qToken != "" {
				authzToken = "Bearer " + qToken
			}
		}

		if authzToken == "" {
			slog.Warn("Authorization header missing in AuthN check",
				slog.String("Path", r.URL.Path),
			)
			errJson, badRequestError := errors.NewBadRequestError("Invalid Request, auth token not present", nil)
			w.WriteHeader(badRequestError.Code)
			w.Write(errJson)
			return
		}

		ctx := r.Context()
		userRole := ctx.Value("role").(string)

		requestPath := r.URL.Path
		isAdminRequest := strings.HasPrefix(requestPath, "/admin")

		if isAdminRequest && userRole != "admin" {
			slog.Warn("Unauthorized admin access attempt",
				slog.String("Role", userRole),
				slog.String("Path", requestPath),
			)
			errJson, unauthorizedError := errors.NewUnauthorizedError(
				"Insufficient permissions to access the resource", nil,
			)
			w.WriteHeader(unauthorizedError.Code)
			w.Write(errJson)
			return
		}

		slog.Info("User authenticated",
			slog.String("Role", userRole),
			slog.String("Path", requestPath),
		)

		next.ServeHTTP(w, r)
	})
}
