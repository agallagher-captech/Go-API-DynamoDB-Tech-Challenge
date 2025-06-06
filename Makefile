SHELL := /bin/bash

.PHONY: swag-init
swag-init:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Generating swagger docs..."
	@swag init \
		--generalInfo "./../../internal/routes/routes.go" \
		--output "cmd/api/docs" \
		--dir "./internal/handlers"
	@swag fmt
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Swagger docs generated"

.PHONY: mock-gen
mock-gen:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Generating mocks..."
	@$(MAKE) LOG MSG_TYPE=debug LOG_MESSAGE="Delete existing mocks"
	@find ./internal -d | grep ^.*mock$$ | xargs rm -rf
	@mockery --all
	@$(MAKE) LOG MSG_TYPE=success LOG_MESSAGE="Mocks generated"

.PHONY: start-web-app
start-web-app:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Starting web app..."
	@$(MAKE) start-database
	@$(MAKE) LOG MSG_TYPE=success LOG_MESSAGE="Started database"
	@go run cmd/api/main.go

.PHONY: stop-web-app
stop-web-app:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Stopping web app..."
	@$(MAKE) stop-database
	@$(MAKE) LOG MSG_TYPE=success LOG_MESSAGE="Stopped database"

.PHONY: start-database
start-database:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Starting database..."
	@docker compose up -d

.PHONY: stop-database
stop-database:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Stopping database..."
	@docker compose down

.PHONY: seed-database
seed-database:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Seeding database..."
	@cd ./dynamodb_seed && /bin/bash ./seed_dynamodb.sh

.PHONE: reset-database
reset-database:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Resetting database..."
	@cd ./dynamodb_seed && /bin/bash ./reset_dynamodb.sh

.PHONY: test
test:
	go test -cover ./internal/**

.PHONY: check-coverage
check-coverage:
	@$(MAKE) LOG MSG_TYPE=info LOG_MESSAGE="Running unit tests and generating coverage report..."
	go test -coverprofile=coverage.out ./internal/service ./internal/config ./internal/database ./cmd/routes ./cmd/api
	go tool cover -html=coverage.out -o coverage.html
	@$(MAKE) LOG MSG_TYPE=warn LOG_MESSAGE="Link to coverage report file: file://$$(PWD)/coverage.html"

.PHONY: view-coverage
view-coverage:
	@open -a "Google Chrome" file://$$(PWD)/coverage.html

LOG:
	@if [ "$(MSG_TYPE)" = "debug" ]; then \
		echo -e "\033[0;37m$(LOG_MESSAGE)\033[0m"; \
	elif [ "$(MSG_TYPE)" = "info" ]; then \
		echo -e "\033[0;36m$(LOG_MESSAGE)\033[0m"; \
	elif [ "$(MSG_TYPE)" = "warn" ]; then \
		echo -e "\033[0;33m$(LOG_MESSAGE)\033[0m"; \
	elif [ "$(MSG_TYPE)" = "success" ]; then \
		echo -e "\033[0;32m$(LOG_MESSAGE)\033[0m✓"; \
	elif [ "$(MSG_TYPE)" = "failure" ]; then \
		echo -e "\033[0;31m$(LOG_MESSAGE)\033[0m"; \
	else \
		echo -e "\033[0;37m$(LOG_MESSAGE)\033[0m"; \
	fi