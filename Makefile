PROJECT_DIR   = $(shell pwd)
PROJECT_BIN   = $(PROJECT_DIR)/bin
STROLTM_UI   = $(PROJECT_DIR)/apps/stroltm/ui
GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint

.install-swagger-client:
	[ -f $(PROJECT_BIN)/swagger-client ] || curl -sSfL "https://github.com/go-swagger/go-swagger/releases/download/v0.30.3/swagger_$(shell sh ./scripts/get_platform.sh)" > $(PROJECT_BIN)/swagger-client && chmod +x $(PROJECT_BIN)/swagger-client

.install-golangci-lint:
	[ -f $(PROJECT_BIN)/golangci-lint ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_BIN) v1.50.0

.install-stroltm-ui-node_modules:
	[ -d $(STROLTM_UI)/node_modules ] || cd $(STROLTM_UI) && yarn --frozen-lockfile --non-interactive -s

all: .install-stroltm-ui-node_modules .install-swagger-client .install-swag

.PHONY: coverage
coverage:
	go test --coverprofile=coverage.strolt.out ./apps/strolt/...
	go tool cover -func=coverage.strolt.out
	rm coverage.strolt.out

##### SWAGGER #####
.install-swag:
	go install github.com/swaggo/swag/cmd/swag@v1.8.7

.swagger-strolt: .install-swag
	cd ./apps/strolt && swag init -g ./internal/api/api.go --parseDependency --output $(PROJECT_DIR)/.swagger/strolt

.swagger-strolt-manager: .install-swag
	cd ./apps/stroltm && swag init -g ./internal/api/api.go --parseDependency --output $(PROJECT_DIR)/.swagger/stroltm

.swagger-strolt-manager-generate-client: .install-swagger-client
	rm -rf ./apps/stroltm/internal/sdk/strolt/generated/client && rm -rf ./apps/stroltm/internal/sdk/strolt/generated/models
	cd ./apps/stroltm/internal/sdk/strolt/generated && $(PROJECT_BIN)/swagger-client generate client -f $(PROJECT_DIR)/.swagger/strolt/swagger.yaml

.swagger-strolt-manager-ui-generate-client: .install-stroltm-ui-node_modules
	cd $(STROLTM_UI) && yarn gen-api

swagger: .swagger-strolt .swagger-strolt-manager-generate-client .swagger-strolt-manager .swagger-strolt-manager-ui-generate-client

##### LINT #####
.lint-strolt: .install-golangci-lint
	cd ./apps/strolt && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-stroltm: .install-golangci-lint
	cd ./apps/stroltm && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-stroltm-ui: .install-stroltm-ui-node_modules
	cd $(STROLTM_UI) && yarn typecheck

.PHONY: lint
lint: .lint-strolt .lint-stroltm .lint-stroltm-ui


##### TEST #####
.test-strolt:
	cd ./apps/strolt && go test $$(go list ./... | grep -v /e2e)

.test-stroltm:
	cd ./apps/stroltm && go test ./...

.PHONY: test
test: .test-strolt .test-stroltm

##### DOCKER #####
.PHONY: docker-strolt
docker-strolt:
	docker build -f ./docker/strolt/Dockerfile --build-arg version=development -t strolt/strolt:development ./

.PHONY: docker-stroltm
docker-stroltm:
	docker build -f ./docker/stroltm/Dockerfile --build-arg version=development -t strolt/stroltm:development ./

.PHONY: docker
docker: docker-strolt

##### E2E TEST #####
.e2e-strolt: docker-strolt
	cd ./apps/strolt && GOFLAGS="-count=1" go test ./e2e

.PHONY: e2e
e2e: .e2e-strolt
