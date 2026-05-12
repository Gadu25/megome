package middleware

import (
	"fmt"
	"net/http"
	"time"

	"megome/internal/services/auth"
	"megome/internal/services/types"
)

func WithAPILogging(
	next http.HandlerFunc,
	apiLogStore types.APIUsageLogStore,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := NewStatusRecorder(w)

		next(recorder, r)

		duration := time.Since(start)

		userID := auth.GetPATUserIDFromContext(r.Context())
		tokenID := auth.GetPATTokenIDFromContext(r.Context())

		log := types.APIUsageLog{
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
