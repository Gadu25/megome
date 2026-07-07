package middleware

import (
	"fmt"
	"megome/internal/domain/apilog"
	"net/http"
	"time"
)

func WithAPILogging(
	next http.HandlerFunc,
	apiLogStore apilog.APIUsageLogStore,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := NewStatusRecorder(w)

		next(recorder, r)

		duration := time.Since(start)

		userID := GetPATUserIDFromContext(r.Context())
		tokenID := GetPATTokenIDFromContext(r.Context())

		log := apilog.APIUsageLog{
			UserID:         userID,
			TokenID:        tokenID,
			Endpoint:       r.URL.Path,
			Method:         r.Method,
			StatusCode:     recorder.StatusCode,
			IPAddress:      r.RemoteAddr,
			UserAgent:      r.UserAgent(),
			ResponseTimeMs: int(duration.Milliseconds()),
		}

		// avoid breaking request flow if logging fails
		err := apiLogStore.Create(log)
		if err != nil {
			fmt.Println("[DEBUG] API LOG CREATE", err)
		}
	}
}
