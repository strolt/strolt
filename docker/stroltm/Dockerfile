# droast global ignore=DF020 reason="manager runs as root by design" expires=2026-12-31
FROM alpine:3.23.7@sha256:ff7f0c727dd796aa36787bfcd9a64bf35a6ce72a73451b3359fde849b6703d63 AS base
WORKDIR /stroltm
# droast ignore=DF052 reason="versions follow the pinned alpine base image" expires=2026-12-31
# hadolint ignore=DL3018
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

# The ui and swagger stages produce architecture-independent artifacts
# (static dist and generated code), so they run on the build host platform;
# node images also ship no linux/arm/v7 variant.
FROM --platform=$BUILDPLATFORM node:26.13.1-trixie-slim AS ui-base

FROM ui-base AS swagger-base
# Datamitsu needs git to identify the workspace and system CA roots for HTTPS downloads.
# droast ignore=DF005 reason="ca-certificates and git are managed by the Debian repository" expires=2026-12-31
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && rm -rf /var/lib/apt/lists/*

FROM swagger-base AS swagger
WORKDIR /github.com/strolt/strolt
COPY ./package.json ./pnpm-workspace.yaml ./pnpm-lock.yaml ./
COPY ./apps/strolt/package.json ./apps/strolt/
COPY ./apps/stroltm/package.json ./apps/stroltm/
COPY ./apps/stroltm/ui/package.json ./apps/stroltm/ui/
COPY ./apps/stroltp/package.json ./apps/stroltp/
COPY ./shared/package.json ./shared/
COPY ./.github/scripts/package.json ./.github/scripts/
# droast ignore=DF031 reason="pnpm is exact-pinned by the workspace packageManager" expires=2026-12-31
RUN npm install --global pnpm@12.19.0 && git init . && pnpm install --frozen-lockfile --filter strolt --ignore-scripts
COPY ./datamitsu.config.js ./Taskfile.yaml ./lefthook.yaml ./
COPY ./scripts/generate-go-sdk.sh ./scripts/
COPY ./.go-swagger/templates ./.go-swagger/templates
COPY ./apps/strolt ./apps/strolt
COPY ./apps/stroltp ./apps/stroltp
COPY ./shared ./shared
RUN pnpm dm init --skip-download --quiet && pnpm dm exec task -- codegen:go -- strolt stroltp && rm -rf .git

FROM ui-base AS ui-deps
WORKDIR /github.com/strolt/strolt
COPY ./package.json ./pnpm-workspace.yaml ./pnpm-lock.yaml ./
COPY ./apps/strolt/package.json ./apps/strolt/
COPY ./apps/stroltm/package.json ./apps/stroltm/
COPY ./apps/stroltm/ui/package.json ./apps/stroltm/ui/
COPY ./apps/stroltp/package.json ./apps/stroltp/
COPY ./shared/package.json ./shared/
COPY ./.github/scripts/package.json ./.github/scripts/
# droast ignore=DF031 reason="pnpm is exact-pinned by the workspace packageManager" expires=2026-12-31
RUN npm install --global pnpm@12.19.0 && pnpm install --frozen-lockfile --filter ./apps/stroltm/ui --ignore-scripts

FROM ui-deps AS ui-builder
WORKDIR /github.com/strolt/strolt/apps/stroltm/ui
COPY ./apps/stroltm/ui .
RUN pnpm build

FROM golang:1.27.3-alpine3.23 AS builder-deps
WORKDIR /go/src/github.com/strolt/strolt/apps/stroltm
COPY ./apps/stroltm/go.mod .
COPY ./apps/stroltm/go.sum .
COPY ./shared /go/src/github.com/strolt/strolt/shared
RUN go mod download

FROM builder-deps AS builder
ARG version
COPY ./apps/stroltm/main.go .
COPY ./apps/stroltm/cmd ./cmd
COPY ./apps/stroltm/internal ./internal
COPY --from=ui-builder /github.com/strolt/strolt/apps/stroltm/ui/dist ./internal/ui/build/dist
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/strolt/generated /go/src/github.com/strolt/strolt/shared/sdk/strolt/generated
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/stroltp/generated /go/src/github.com/strolt/strolt/shared/sdk/stroltp/generated
RUN go build -o stroltm -ldflags "-X github.com/strolt/strolt/apps/stroltm/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA
ARG RELEASE_VERSION=development
ARG RELEASE_RUN_ID=0
ARG RELEASE_ATTEMPT=0
ARG RELEASE_WORKFLOW_SHA=development
ARG RELEASE_CHANNEL=development

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
			org.opencontainers.image.description="Strolt manager" \
			org.opencontainers.image.licenses="MIT" \
			org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
			org.opencontainers.image.title="Stroltm" \
            org.opencontainers.image.revision="${GITHUB_SHA}" \
            org.opencontainers.image.version="${RELEASE_VERSION}" \
            com.strolt.release.run="${RELEASE_RUN_ID}" \
            com.strolt.release.attempt="${RELEASE_ATTEMPT}" \
            com.strolt.release.workflow="${RELEASE_WORKFLOW_SHA}" \
            com.strolt.release.channel="${RELEASE_CHANNEL}" \
            com.strolt.release.product="stroltm"

COPY --from=builder /go/src/github.com/strolt/strolt/apps/stroltm/stroltm ./bin/stroltm

EXPOSE 8080

ENTRYPOINT ["/stroltm/bin/stroltm", "--config", "/stroltm/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "--fail", "http://localhost:8080/api/v1/ping"]

CMD ["--json"]
