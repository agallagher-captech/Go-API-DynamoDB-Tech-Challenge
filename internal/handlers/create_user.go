package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agallagher-captech/blog/internal/models"
)

func HandleCreateUser(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "handling create user request")

		// Set the status code to 200 OK
		w.WriteHeader(http.StatusOK)

		// Get user from request body
		var user models.User
        if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
            http.Error(w, fmt.Sprintf("Failed to decode request body: %v", err), http.StatusBadRequest)
            return
        }

        user.SK = "PROFILE"
        user.GSI1PK = "USER"

		// Write the response body, simply echo the ID back out
		_, err := w.Write([]byte(fmt.Sprintf("User ID: %s", user.ID)))
		if err != nil {
			// Handle error if response writing fails
			logger.ErrorContext(r.Context(), "failed to write response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}