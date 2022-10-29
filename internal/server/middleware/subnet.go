package middleware

import (
	"errors"
	"net"
	"net/http"

	"metrics/internal/server/responses"
)

func NewSubNetHandle(TrustedSubNetStr string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipStr := r.Header.Get("X-Real-IP")
			response := responses.NewUpdateMetricResponse()

			clientIP := net.ParseIP(ipStr)
			if clientIP == nil {
				http.Error(w, response.SetStatusError(errors.New("unknown client IP")).GetJSONString(), http.StatusForbidden)
				return
			}

			_, TrustedSubNet, _ := net.ParseCIDR(TrustedSubNetStr)

			if !TrustedSubNet.Contains(clientIP) {
				http.Error(w, response.SetStatusError(errors.New("client IP is not in trusted subnet")).GetJSONString(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
