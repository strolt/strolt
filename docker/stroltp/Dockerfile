# droast global ignore=DF020 reason="proxy runs as root by design" expires=2026-12-31
FROM alpine:3.23.4 AS base
WORKDIR /stroltp
# droast ignore=DF052 reason="versions follow the pinned alpine base image" expires=2026-12-31
# hadolint ignore=DL3018
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

FROM golang:1.26.4-alpine3.23 AS swagger-base
# droast ignore=DF052 reason="build stage follows the pinned alpine base image" expires=2026-12-31
# hadolint ignore=DL3018
RUN apk --update add --no-cache make curl

FROM swagger-base AS swagger
WORKDIR /github.com/strolt/strolt
# droast ignore=DF007 reason="swagger codegen needs full repo context across apps and shared" expires=2026-12-31
COPY . .
RUN make .swagger-strolt && make .swagger-shared-generate-client-strolt

FROM golang:1.26.4-alpine3.23 AS builder
ARG version
WORKDIR /go/src/github.com/strolt/strolt/apps/stroltp
COPY ./apps/stroltp/go.mod .
COPY ./apps/stroltp/go.sum .
COPY ./shared /go/src/github.com/strolt/strolt/shared
RUN go mod download
COPY ./apps/stroltp/main.go .
COPY ./apps/stroltp/cmd ./cmd
COPY ./apps/stroltp/internal ./internal
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/strolt/generated /go/src/github.com/strolt/strolt/shared/sdk/strolt/generated
RUN go build -o stroltp -ldflags "-X github.com/strolt/strolt/apps/stroltp/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA
ENV STROLTP_PATH_DATA=/stroltp

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
			org.opencontainers.image.description="strolt proxy" \
			org.opencontainers.image.licenses="MIT" \
			org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
			org.opencontainers.image.title="stroltp" \
			org.opencontainers.image.revision="${GITHUB_SHA}"

COPY --from=builder /go/src/github.com/strolt/strolt/apps/stroltp/stroltp ./bin/stroltp

EXPOSE 8080

ENTRYPOINT ["/stroltp/bin/stroltp", "--config", "/stroltp/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "--fail", "http://localhost:8080/api/v1/ping"]

CMD ["start", "--json"]
