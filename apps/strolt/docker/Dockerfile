FROM alpine:3.17.0 AS base
WORKDIR /strolt
COPY ./docker/scripts scripts
RUN sh ./scripts/install.sh && rm -rf ./scripts

FROM golang:1.19.0-alpine3.16@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 AS builder
ARG version
WORKDIR /go/src/strolt/strolt/strolt
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY ./main.go .
COPY ./cmd ./cmd
COPY ./internal ./internal
RUN go build -o strolt -ldflags "-X github.com/strolt/strolt/apps/strolt/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
      org.opencontainers.image.description="Strolt backup engine" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
      org.opencontainers.image.title="Strolt" \
      org.opencontainers.image.revision="${GITHUB_SHA}"

COPY --from=builder /go/src/strolt/strolt/strolt/strolt ./bin/strolt

ENTRYPOINT ["/strolt/bin/strolt", "--config", "/strolt/config"]

HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/api/v1/ping || exit 1

CMD ["--json"]
