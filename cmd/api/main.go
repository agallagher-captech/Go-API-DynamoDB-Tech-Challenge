package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/agallagher-captech/blog/internal/configuration"
	"github.com/agallagher-captech/blog/internal/middleware"
	"github.com/agallagher-captech/blog/internal/routes"
	"github.com/agallagher-captech/blog/internal/services"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server encountered an error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	// Load and validate environment configuration
	cfg, err := configuration.New()
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load configuration: %w", err)
	}

	// Create a structured logger, which will print logs in json format to the
	// writer we specify.
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// connect to dynamoDB
	logger.InfoContext(ctx, "connecting to DynamoDB")
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load configuration: %w", err)
	}

	client := dynamodb.NewFromConfig(awsCfg, func(options *dynamodb.Options) {
		options.BaseEndpoint = aws.String(cfg.DynamoEndpoint)
	})

	// list all tables in db (we will delete this later)
	result, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return fmt.Errorf("[in main.run] failed to list tables: %w", err)
	}

	fmt.Println("Tables:")
	for _, tableName := range result.TableNames {
		fmt.Printf("* %s\n", tableName)
	}

	// Create a new users service
	usersService := services.NewUsersService(logger, client)

	// Create a serve mux to act as our route multiplexer
	mux := http.NewServeMux()

	// Add our routes to the mux
	// Add our routes to the mux
	routes.AddRoutes(
		mux,
		logger,
		usersService,
		fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port),
	)
	// Wrap the mux with middleware
	wrappedMux := middleware.Logger(logger)(mux)

	// Create a new http server with our mux as the handler
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: wrappedMux,
	}

	errChan := make(chan error)

	// Server run context
	ctx, done := context.WithCancel(ctx)
	defer done()

	// Handle graceful shutdown with go routine on SIGINT
	go func() {
		// create a channel to listen for SIGINT and then block until it is received
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig

		logger.DebugContext(ctx, "Received SIGINT, shutting down server")

		// Create a context with a timeout to allow the server to shut down gracefully
		ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ShutdownTimout)*time.Second)
		defer cancel()

		// Shutdown the server. If an error occurs, send it to the error channel
		if err = httpServer.Shutdown(ctx); err != nil {
			errChan <- fmt.Errorf("[in main.run] failed to shutdown http server: %w", err)
			return
		}

		// Close the idle connections channel, unblocking `run()`
		done()
	}()

	// Start the http server
	//
	// once httpServer.Shutdown is called, it will always return a
	// http.ErrServerClosed error and we don't care about that error.
	logger.InfoContext(ctx, "listening", slog.String("address", httpServer.Addr))
	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("[in main.run] failed to listen and serve: %w", err)
	}

	// block until the server is shut down or an error occurs
	select {
	case err = <-errChan:
		return err
	case <-ctx.Done():
		logger.InfoContext(ctx, "server shutdown complete")
		return nil
	}
}
