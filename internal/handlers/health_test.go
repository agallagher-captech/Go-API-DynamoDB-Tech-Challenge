package handlers

import (
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleHealthCheck(t *testing.T) {
	tests := map[string]struct {
		wantStatus int
		wantBody   string
	}{
		"happy path": {
			wantStatus: 200,
			wantBody:   `{"status":"ok"}`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create a new request
			req := httptest.NewRequest("GET", "/health", nil)

			// Create a new response recorder
			rec := httptest.NewRecorder()

			// Create a new logger
			logger := slog.Default()

			// Call the handler
			HandleHealthCheck(logger)(rec, req)

			// Check the status code
			assert.Equal(t, tc.wantStatus, rec.Code, "status code mismatch")

			// Check the body
			assert.JSONEq(t, tc.wantBody, strings.Trim(rec.Body.String(), "\n"), "body mismatch")
		})
	}
}
