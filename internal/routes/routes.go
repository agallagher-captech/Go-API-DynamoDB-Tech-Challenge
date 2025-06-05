package routes

import (
	"log/slog"
	"net/http"

	_ "github.com/agallagher-captech/blog/cmd/api/docs"
	"github.com/agallagher-captech/blog/internal/handlers"
	"github.com/agallagher-captech/blog/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
)

// AddRoutes adds all routes to the provided mux.
//
//	@title						Blog Service API
//	@version					1.0
//	@description				Practice Go API using the Standard Library and DynamoDB
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				http://www.swagger.io/support
//	@contact.email				support@swagger.io
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@host						localhost:8080
//	@BasePath					/api
//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
func AddRoutes(mux *http.ServeMux, logger *slog.Logger, usersService *services.UsersService, baseURL string) {
	// Swagger docs
	mux.Handle(
		"GET /swagger/",
		httpSwagger.Handler(httpSwagger.URL(baseURL+"/swagger/doc.json")),
	)
	logger.Info("Swagger running", slog.String("url", baseURL+"/swagger/index.html"))

	// Health check
	mux.Handle("GET /api/health", handlers.HandleHealthCheck(logger))

	// Read a user
	mux.Handle("GET /api/users/{id}", handlers.HandleReadUser(logger, usersService))

	// Create a user
	mux.Handle("/api/users", handlers.HandleCreateUser(logger, usersService))
}
