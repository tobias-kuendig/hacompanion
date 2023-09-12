package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondError returns a JSON error response.
//
//nolint:errcheck
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

// Because Home Assistant errors without a rateLimits response and
// there is no rate limiting implemented on our side, we return a
// dummy response with RespondSuccess.
//
//nolint:errcheck
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
