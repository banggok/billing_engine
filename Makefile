# Load environment variables from .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

GOCMD=go
GINKGO=ginkgo
COVERAGE_OUT=coverage.out
NETWORK_NAME=shared_services

# Check if Docker network exists and containers are running
.PHONY: check-containers
check-containers:
	# Check if the Docker network exists, if not, create it
	@if ! docker network inspect $(NETWORK_NAME) >/dev/null 2>&1; then \
		echo "Network $(NETWORK_NAME) not found. Creating..."; \
		docker network create $(NETWORK_NAME); \
	fi
	# Check if containers are running, if not, start them
	@docker-compose ps | grep "Up" > /dev/null || docker-compose up -d

# Default target: Run the service with Docker Compose
.PHONY: run
run: check-containers
	@echo "Running the Inventory Management service..."
	$(GOCMD) run ./cmd/api

# Migration commands: Run migrations using soda (change soda to your migration tool if different)
.PHONY: migrate
migrate: check-containers
	@echo "Running migrations..."
	soda migrate

.PHONY: rollback
rollback: check-containers
	@echo "Rolling back migrations..."
	soda migrate down

# Run the E2E tests with Docker Compose and coverage for all folders
.PHONY: test
test: check-containers
	@echo "Running E2E tests with coverage..."
	$(GINKGO) -r -race -cover -coverpkg=./... -coverprofile=$(COVERAGE_OUT)

# Open the coverage report in HTML format
.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html
	@echo "Opening coverage report..."
	open coverage.html # On Linux, you might need to use xdg-open

# Performing code analysis...
.PHONY: analysis
analysis: test
	@echo "Performing code analysis..."
	golangci-lint run
	@echo "Running gosec security checks..."
	gosec -severity medium -confidence medium ./... 
	@if [ $$? -ne 0 ]; then \
	  echo "gosec detected security issues! Failing the build."; \
	  exit 1; \
	fi

# Run SonarScanner after the tests and coverage generation
.PHONY: sonar
sonar: analysis
	@echo "Loading environment variables and running SonarScanner for code analysis..."
	sonar-scanner -Dsonar.host.url=$(SONAR_HOST_URL) -Dsonar.token=$(SONAR_TOKEN)

# Clean up: Remove the coverage report file
.PHONY: clean
clean:
	@echo "Cleaning up coverage files..."
	rm -f $(COVERAGE_OUT) coverage.html

# Run all: Start Docker containers, run migrations, the service, tests, coverage, and sonar analysis
.PHONY: all
all: migrate run test coverage sonar

# Stop Docker Compose services
.PHONY: stop
stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Start the server locally, ensuring migrations are run first
.PHONY: start-local
start-local: migrate
	$(GOCMD) run cmd/api/main.go
