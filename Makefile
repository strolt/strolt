PROJECT_DIR   = $(shell pwd)
PROJECT_BIN   = $(PROJECT_DIR)/bin
GOLANGCI_LINT = $(PROJECT_BIN)/golangci-lint

DOCKER_IMAGE   ?= strolt/strolt
STROLT_VERSION ?= 0.0.8-alpha.8-nightly
DOCKER_REF     := $(DOCKER_IMAGE):$(STROLT_VERSION)
DOCKER_E2E_REF := $(DOCKER_IMAGE):e2e

.PHONY: .install-swagger-client
.install-swagger-client:
	[ -f $(PROJECT_BIN)/swagger-client ] || curl -sSfL "https://github.com/go-swagger/go-swagger/releases/download/v0.30.3/swagger_$(shell sh ./scripts/get_platform.sh)" > $(PROJECT_BIN)/swagger-client && chmod +x $(PROJECT_BIN)/swagger-client

.PHONY: .install-linter
.install-linter:
	[ -f $(PROJECT_BIN)/golangci-lint ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_BIN) v1.50.0

.PHONY: .install-swag
.install-swag:
	go install github.com/swaggo/swag/cmd/swag@v1.8.7

.PHONY: swagger-strolt
swagger-strolt: .install-swag
	cd ./apps/strolt && swag init -g ./internal/api/api.go --output $(PROJECT_DIR)/.swagger/strolt

.PHONY: swagger-strolt-manager
swagger-strolt-manager: .install-swag
	cd ./apps/stroltm && swag init -g ./internal/api/api.go --output $(PROJECT_DIR)/.swagger/stroltm

.PHONY: swagger-generate-client
swagger-generate-client: .install-swagger-client
	rm -rf ./apps/stroltm/internal/sdk/strolt/generated/client && rm -rf ./apps/stroltm/internal/sdk/strolt/generated/models
	cd ./apps/stroltm/internal/sdk/strolt/generated && $(PROJECT_BIN)/swagger-client generate client -f $(PROJECT_DIR)/.swagger/strolt/swagger.yaml

.PHONY: swagger
swagger: swagger-strolt swagger-generate-client swagger-strolt-manager

.PHONY: lint-strolt
lint-strolt: .install-linter
	cd ./apps/strolt && $(GOLANGCI_LINT) run ./... --config=${PROJECT_DIR}/.golangci.yml

.PHONY: lint-fast-strolt
lint-fast-strolt: .install-linter
	cd ./apps/strolt && $(GOLANGCI_LINT) run ./... --fast --config=${PROJECT_DIR}/.golangci.yml

.PHONY: lint-stroltm
lint-stroltm: .install-linter
	cd ./apps/stroltm && $(GOLANGCI_LINT) run ./... --config=${PROJECT_DIR}/.golangci.yml

.PHONY: lint-fast-stroltm
lint-fast-stroltm: .install-linter
	cd ./apps/stroltm && $(GOLANGCI_LINT) run ./... --fast --config=${PROJECT_DIR}/.golangci.yml

.PHONY: lint
lint: lint-strolt lint-stroltm

.PHONY: lint-fast
lint-fast: lint-fast-strolt lint-fast-stroltm

.PHONY: test
test:
	go test ./apps/strolt/...

.PHONY: coverage
coverage:
	go test --coverprofile=coverage.strolt.out ./apps/strolt/...
	go tool cover -func=coverage.strolt.out
	rm coverage.strolt.out

.PHONY: e2e-strolt
e2e-strolt:
	docker build -f ./apps/strolt/docker/Dockerfile --build-arg STROLT_VERSION=e2e -t $(DOCKER_E2E_REF) ./apps/strolt
	cd ./apps/strolt && GOFLAGS="-count=1" go test ./e2e

.PHONY: e2e
e2e: e2e-strolt






q: .install-swagger-client
	@echo "$(shell sh ./scripts/get_platform.sh)"
	@echo "https://github.com/go-swagger/go-swagger/releases/download/v0.30.3/swagger_$(shell sh ./scripts/get_platform.sh)"


# download_url=$(curl -s https://api.github.com/repos/go-swagger/go-swagger/releases/latest | \
#   jq -r '.assets[] | select(.name | contains("'"$(uname | tr '[:upper:]' '[:lower:]')"'_amd64")) | .browser_download_url')
# curl -o /usr/local/bin/swagger -L'#' "$download_url"
# chmod +x /usr/local/bin/swagger
