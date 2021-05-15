package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondError return a JSON error response.
// nolint:errcheck
func RespondError(w http.ResponseWriter, error string, status int) {
	var resp struct {
		Error string `json:"errorMessage"`
	}
	resp.Error = error
	b, err := json.Marshal(resp)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, string(b))))
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(status)
	w.Write(b)
}

// Home Assistant errors without a rateLimits response.
// There is no rate limiting implemented on our side so we return a dummy response.
// nolint:errcheck
func RespondSuccess(w http.ResponseWriter) {
	w.Write([]byte(`
		{
			"rateLimits": {
				"successful": 1,
				"errors": 0,
				"maximum": 150,
				"resetsAt": "2019-04-08T00:00:00.000Z"
			}
		}
	`))
}
