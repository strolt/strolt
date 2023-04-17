FROM alpine:3.17.3 AS base
WORKDIR /strolt
COPY ./docker/strolt/scripts scripts
RUN sh ./scripts/install.sh && rm -rf ./scripts
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

FROM golang:1.20.3-alpine3.17 AS builder
ARG version
WORKDIR /go/src/github.com/strolt/strolt/apps/strolt
COPY ./apps/strolt/go.mod .
COPY ./apps/strolt/go.sum .
COPY ./shared /go/src/github.com/strolt/strolt/shared
RUN go mod download
COPY ./apps/strolt/main.go .
COPY ./apps/strolt/cmd ./cmd
COPY ./apps/strolt/internal ./internal
RUN go build -o strolt -ldflags "-X github.com/strolt/strolt/apps/strolt/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA
ENV STROLT_PATH_DATA=/strolt

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
      org.opencontainers.image.description="Strolt backup engine" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
      org.opencontainers.image.title="Strolt" \
      org.opencontainers.image.revision="${GITHUB_SHA}"

COPY --from=builder /go/src/github.com/strolt/strolt/apps/strolt/strolt ./bin/strolt

ENTRYPOINT ["/strolt/bin/strolt", "--config", "/strolt/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/api/v1/ping || exit 1

CMD ["start", "--json"]
