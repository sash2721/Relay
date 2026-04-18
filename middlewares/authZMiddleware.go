package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sash2721/Relay/errors"
	"github.com/sash2721/Relay/utils"
)

func AuthZMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authzToken := r.Header.Get("Authorization")

		// Also check query param for SSE connections (EventSource can't set headers)
		if authzToken == "" {
			if qToken := r.URL.Query().Get("token"); qToken != "" {
				authzToken = "Bearer " + qToken
			}
		}

		if authzToken == "" {
			slog.Warn("Authorization header missing",
				slog.String("Path", r.URL.Path),
			)
			errJson, badRequestError := errors.NewBadRequestError("Invalid Request, auth token not present", nil)
			w.WriteHeader(badRequestError.Code)
			w.Write(errJson)
			return
		}

		tokenString := strings.TrimPrefix(authzToken, "Bearer ")

		userInfo, err, errJson, errorCode := utils.ValidateToken(tokenString)

		if err != nil {
			if errorCode == http.StatusInternalServerError {
				slog.Error("Token validation failed due to internal error",
					slog.String("Path", r.URL.Path),
				)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(errJson)
				return
			} else if errorCode == http.StatusBadRequest {
				slog.Warn("Invalid token provided",
					slog.String("Path", r.URL.Path),
				)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(errJson)
				return
			}
		}

		slog.Info("JWT token validated successfully",
			slog.String("UserID", userInfo.UserID),
			slog.String("Role", userInfo.Role),
		)

		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", userInfo.UserID)
		ctx = context.WithValue(ctx, "email", userInfo.Email)
		ctx = context.WithValue(ctx, "role", userInfo.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
