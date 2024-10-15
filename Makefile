setup: fmt-install lint-install dep ## Setup the project

fmt-install:
	go install github.com/daixiang0/gci@v0.13.4
	go install mvdan.cc/gofumpt@v0.6.0

fmt: ## Format the code
	gci write -s standard -s default -s localmodule -s blank -s dot --skip-generated --custom-order --skip-vendor cmd internal pkg
	gofumpt -l -w -extra cmd internal pkg

lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

lint: dep fmt ## Lint the code
	golangci-lint run ./...

test: ## Run the tests
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

dep: ## Sync dependencies
	go mod tidy && go mod verify

env-start: ## Start the local env
	docker-compose -f temporal-server/docker-compose.yml up -d

env-stop: ## Stop the local env
	docker-compose -f temporal-server/docker-compose.yml down

env-restart: env-stop env-start ## Restart the local env


# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
