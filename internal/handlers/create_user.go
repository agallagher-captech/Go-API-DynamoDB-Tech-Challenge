package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/agallagher-captech/blog/internal/models"
	"github.com/agallagher-captech/blog/internal/services"
	"github.com/google/uuid"
)

// createUserRequest represents the input model for creating a user.
type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Valid checks the createUserRequest for any problems.
func (r createUserRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	} else if len(r.Name) < 2 {
		problems["name"] = "name must be at least 2 characters"
	}
	if strings.TrimSpace(r.Email) == "" {
		problems["email"] = "email is required"
	} else if !isValidEmail(r.Email) {
		problems["email"] = "invalid email format"
	}
	if len(r.Password) < 8 {
		problems["password"] = "password must be at least 8 characters"
	}

	return problems
}

func isValidEmail(email string) bool {
	// A simple regex for email validation
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// userCreator represents a type capable of creating a user in storage.
type userCreator interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

// HandleCreateUser returns an http.Handler that creates a new user.
//
//	@Summary		Create User
//	@Description	Create a new user in the system
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		createUserRequest	true	"User creation request"
//	@Success		201		{object}	userResponse
//	@Failure		400		{object}	map[string]string	"Validation error(s)"
//	@Failure		409		{object}	string				"User already exists"
//	@Failure		500		{object}	string				"Internal server error"
//	@Router			/users [post]
func HandleCreateUser(logger *slog.Logger, userCreator userCreator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "handling create user request")

		// decode and validate request
		req, problems, err := decodeValid[createUserRequest](ctx, r)
		if err != nil {
			logger.ErrorContext(ctx, "invalid create user request", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if problems != nil {
				_ = json.NewEncoder(w).Encode(problems)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}

		// Set the status code to 200 OK
		w.WriteHeader(http.StatusOK)

		// Get user from request body
		var user models.User
		user = models.User{
			ID:       models.UUID{UUID: uuid.New()},
			Name:     req.Name,
			Email:    req.Email,
			Password: req.Password,
		}

		user.SK = "PROFILE"
		user.GSI1PK = "USER"
		user.GSI1SK = "USER#" + user.ID.String() // GSI1SK must be unique for each user

		// Create the user
		createdUser, err := userCreator.CreateUser(ctx, user)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrAlreadyExists):
				logger.ErrorContext(ctx, "user already exists")
				http.Error(w, "User already exists", http.StatusConflict)
				return
			default:
				logger.ErrorContext(ctx, "failed to create user", slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Convert our models.User domain model into a response model.
		response := userResponse{
			ID:       createdUser.ID.UUID,
			Name:     createdUser.Name,
			Email:    createdUser.Email,
			Password: createdUser.Password, // Note: Password should not be returned in production
		}

		// Encode the response model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201 Created
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}
