PROJECT_DIR   = $(shell pwd)
PROJECT_BIN   = $(PROJECT_DIR)/bin
STROLTM_UI   = $(PROJECT_DIR)/apps/stroltm/ui
GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint
GO_SWAGGER_TEMPLATES = $(PROJECT_DIR)/.go-swagger/templates


.install-swagger-client:
	[ -f $(PROJECT_BIN)/swagger-client ] || curl -sSfL "https://github.com/go-swagger/go-swagger/releases/download/v0.30.4/swagger_$(shell sh ./scripts/get_platform.sh)" > $(PROJECT_BIN)/swagger-client && chmod +x $(PROJECT_BIN)/swagger-client

.install-golangci-lint:
	[ -f $(PROJECT_BIN)/golangci-lint ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_BIN) v1.52.2

.install-stroltm-ui-node_modules:
	[ -d $(STROLTM_UI)/node_modules ] || (cd $(STROLTM_UI) && pnpm install --frozen-lockfile)

all: .install-stroltm-ui-node_modules .install-swagger-client .install-swag

.PHONY: coverage
coverage:
	go test --coverprofile=coverage.strolt.out ./apps/strolt/...
	go tool cover -func=coverage.strolt.out
	rm coverage.strolt.out

##### SWAGGER #####
.clear-sdk:
	rm -rf .swagger
	rm -rf ./shared/sdk/strolt/generated/strolt_client && rm -rf ./shared/sdk/strolt/generated/strolt_models
	rm -rf ./shared/sdk/stroltp/generated/stroltp_client && rm -rf ./shared/sdk/stroltp/generated/stroltp_models

.install-swag:
	go install github.com/swaggo/swag/cmd/swag@v1.16.6

.swagger-strolt: .install-swag
	cd ./apps/strolt && swag init -g ./internal/api/api.go -q --parseDependency --output $(PROJECT_DIR)/.swagger/strolt

.swagger-stroltm: .install-swag
	cd ./apps/stroltm && swag init -g ./internal/api/api.go -q --parseDependency --output $(PROJECT_DIR)/.swagger/stroltm

.swagger-stroltp: .install-swag
	cd ./apps/stroltp && swag init -g ./internal/api/api.go -q --parseDependency --output $(PROJECT_DIR)/.swagger/stroltp

.swagger-shared-generate-client-strolt: .install-swagger-client
	rm -rf ./shared/sdk/strolt/generated/strolt_client && rm -rf ./shared/sdk/strolt/generated/strolt_models
	cd ./shared/sdk/strolt/generated && $(PROJECT_BIN)/swagger-client generate client -q -c stroltClient -m stroltModels -f $(PROJECT_DIR)/.swagger/strolt/swagger.yaml --template-dir=${GO_SWAGGER_TEMPLATES} --allow-template-override

.swagger-shared-generate-client-stroltp: .install-swagger-client
	rm -rf ./shared/sdk/stroltp/generated/stroltp_client && rm -rf ./shared/sdk/stroltp/generated/stroltp_models
	cd ./shared/sdk/stroltp/generated && $(PROJECT_BIN)/swagger-client generate client -q -c stroltpClient -m stroltpModels -f $(PROJECT_DIR)/.swagger/stroltp/swagger.yaml --template-dir=${GO_SWAGGER_TEMPLATES} --allow-template-override


.swagger-stroltm-ui-generate-client: .install-stroltm-ui-node_modules
	cd $(STROLTM_UI) && pnpm gen-api

swagger: \
	.clear-sdk \
	.swagger-strolt \
	.swagger-shared-generate-client-strolt \
	.swagger-stroltp \
	.swagger-shared-generate-client-stroltp \
	.swagger-stroltm \
	.swagger-stroltm-ui-generate-client


##### LINT #####
.lint-strolt: .install-golangci-lint
	cd ./apps/strolt && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-stroltm: .install-golangci-lint
	cd ./apps/stroltm && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-stroltp: .install-golangci-lint
	cd ./apps/stroltp && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-shared: .install-golangci-lint
	cd ./shared && $(GOLANGCI_LINT) run ./... --fix --config=${PROJECT_DIR}/.golangci.yml

.lint-stroltm-ui: .install-stroltm-ui-node_modules
	cd $(STROLTM_UI) && pnpm typecheck

.PHONY: lint
lint: .lint-shared .lint-strolt .lint-stroltp .lint-stroltm .lint-stroltm-ui


##### TEST #####
.test-strolt:
	cd ./apps/strolt && go test $$(go list ./... | grep -v /e2e)

.test-stroltm:
	cd ./apps/stroltm && go test ./...

.test-stroltp:
	cd ./apps/stroltp && go test ./...

.PHONY: test
test: .test-strolt .test-stroltp .test-stroltm

##### DOCKER #####
.PHONY: docker-strolt
docker-strolt:
	docker build -f ./docker/strolt/Dockerfile --build-arg version=development -t strolt/strolt:development ./

.PHONY: docker-stroltm
docker-stroltm:
	docker build -f ./docker/stroltm/Dockerfile --build-arg version=development -t strolt/stroltm:development ./

.PHONY: docker-stroltp
docker-stroltp:
	docker build -f ./docker/stroltp/Dockerfile --build-arg version=development -t strolt/stroltp:development ./

.PHONY: docker
docker: docker-strolt docker-stroltp docker-stroltm

##### E2E TEST #####
.e2e-strolt: docker-strolt
	cd ./apps/strolt && GOFLAGS="-count=1" go test ./e2e

.PHONY: e2e
e2e: .e2e-strolt
