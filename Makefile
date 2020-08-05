PROJECT_NAME := "bitwarden-alfred-workflow"
PKG := "github.com/blacs30/$(PROJECT_NAME)"
GO111MODULE=on
.EXPORT_ALL_VARIABLES:
.PHONY: all dep lint vet test test-coverage build clean

all: build
dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint run --timeout 3m

vet: ## Run go vet
	@go vet

test: ## Run unittests
	@go test -short

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic
	@cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -i -o workflow/$(PROJECT_NAME)

clean: ## Remove previous build
	@rm -f workflow/$(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-hooks:
	@mkdir -p .git/hooks
	@cp .github/hooks/* .git/hooks
	@chmod +x .git/hooks/*
