FROM alpine:3.17.0@sha256:c0d488a800e4127c334ad20d61d7bc21b4097540327217dfab52262adc02380c AS base
WORKDIR /strolt
COPY ./docker/strolt/scripts scripts
RUN sh ./scripts/install.sh && rm -rf ./scripts
RUN apk --update add --no-cache curl=7.87.0-r1

FROM golang:1.19.0-alpine3.16@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 AS builder
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
