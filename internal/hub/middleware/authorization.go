package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"

	"github.com/avozda/global-latency-tracker/api"

	log "github.com/sirupsen/logrus"
)

var (
	APIKeyRequiredError = errors.New("API key is required")
	InvalidAPIKeyError  = errors.New("Invalid API key")
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api_key := os.Getenv("API_KEY")
		if api_key == "" {
			log.Error(APIKeyRequiredError)
			api.RequestErrorHandler(w, APIKeyRequiredError)
			return
		}

		if subtle.ConstantTimeCompare([]byte(api_key), []byte(r.Header.Get("X-API-Key"))) != 1 {
			log.Error(InvalidAPIKeyError)
			api.RequestErrorHandler(w, InvalidAPIKeyError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
