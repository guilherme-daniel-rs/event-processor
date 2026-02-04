
APP_NAME = event-processor
TEST_DIR = ./internal/...
COVERAGE_DIR = ./coverage
COVERAGE_FILE = cover.out

GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOCOVER = $(GOCMD) tool cover
BINARY_NAME = $(APP_NAME)
BINARY_UNIX = $(APP_NAME)_unix

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/$(BINARY_NAME)/main.go
	$(GOBUILD) -o bin/$(BINARY_UNIX) cmd/$(BINARY_UNIX)/main.go

test:
	$(GOTEST) $(TEST_DIR)

coverage:
	mkdir $(COVERAGE_DIR) -p && $(GOTEST) -v $(TEST_DIR) -coverprofile=$(COVERAGE_DIR)/$(COVERAGE_FILE) && $(GOCOVER) -html=$(COVERAGE_DIR)/$(COVERAGE_FILE)

run:
	@echo "Starting worker service..."
	go run cmd/worker/main.go

send-events:
	@echo "Sending events..."
	go run cmd/send-events/main.go $(ARGS)
