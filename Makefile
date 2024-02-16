.DEFAULT_GOAL := help

.PHONY: help
help: Makefile ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n\n    make \033[36m<target>\033[0m [VARIABLE=value...]\n\nTargets:\n\n"}; {printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@grep -E '^(override )?[a-zA-Z_-]+ \??\+?= .*?## .*$$' $< | sort | awk 'BEGIN {FS = " \\??\\+?= .*?## "; printf "\nVariables:\n\n"}; {gsub(/override /, "", $$1); printf "    \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: unit-test
unit-test:
	go test ./... -race -coverprofile cmd-cover.out

.PHONY: clean-go-cache
clean-go-cache: ## Clean go cache
	@go clean -modcache

.PHONY: deps
deps: ## Add missing and remove unused modules, verify deps and download them to local cache
	@go mod tidy && go mod verify && go mod download

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against code
	docker run --pull always --rm -v $(shell pwd):/telemetry-exporter -w /telemetry-exporter -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run

.PHONY: dev-all
dev-all: deps fmt vet lint unit-test ## Run all the development checks

.PHONY: generate
generate: ## Run go generate
	go generate ./...
