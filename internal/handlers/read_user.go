package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/agallagher-captech/blog/internal/models"
	"github.com/agallagher-captech/blog/internal/services"
	"github.com/google/uuid"
)

// userReader represents a type capable of reading a user from storage and
// returning it or an error.
type userReader interface {
	ReadUser(ctx context.Context, id uuid.UUID) (models.User, error)
}

// HandleReadUser returns an http.Handler that reads a user from storage.
//
//	@Summary		Read User
//	@Description	Read User by ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string	true	"User ID"
//	@Success		200				{object}	userResponse
//	@Failure		400				{object}	string
//	@Failure		404				{object}	string
//	@Failure		500				{object}	string
//	@Router			/users/{id}  	[GET]
func HandleReadUser(logger *slog.Logger, userReader userReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "handling read user request")

		// Set the status code to 200 OK
		w.WriteHeader(http.StatusOK)

		idStr := r.PathValue("id")

		// Convert the ID from string to a UUID
		id, err := uuid.Parse(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)

			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Read the user
		user, err := userReader.ReadUser(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNotFound):
				logger.ErrorContext(ctx, "user not found")
				http.Error(w, "User not found", http.StatusNotFound)

			default:
				logger.ErrorContext(
					ctx,
					"failed to read user",
					slog.String("error", err.Error()),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		// Convert our models.User domain model into a response model.
		response := userResponse{
			ID:       user.ID.UUID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		}

		// Encode the response model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.String("error", err.Error()),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
