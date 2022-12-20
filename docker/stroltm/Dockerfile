FROM alpine:3.17.0 AS base
WORKDIR /stroltm

FROM node:18.12.1-alpine3.16 as ui-base
RUN apk add --no-cache openjdk17=17.0.5_p8-r0

FROM golang:1.19.0-alpine3.16@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 AS swagger-base
RUN apk add --no-cache make=4.3-r0 curl=7.83.1-r4

FROM ui-base as ui-deps
WORKDIR /github.com/strolt/strolt/apps/stroltm/ui
COPY ./apps/stroltm/ui/package.json .
COPY ./apps/stroltm/ui/yarn.lock .
RUN yarn install --frozen-lockfile --prefer-offline --ignore-scripts --network-timeout 1000000 && yarn cache clean

FROM swagger-base AS swagger
WORKDIR /github.com/strolt/strolt
COPY . .
RUN make .swagger-strolt && make .swagger-strolt-manager-generate-client && make .swagger-strolt-manager

FROM ui-deps as ui-builder
COPY ./apps/stroltm/ui .
COPY --from=swagger /github.com/strolt/strolt/.swagger /github.com/strolt/strolt/.swagger
RUN yarn gen-api && yarn build

FROM golang:1.19.0-alpine3.16@sha256:70df3b8f9f099da7f60f0b32480015165e3d0b51bfacf9e255b59f3dd6bd2828 AS builder
ARG version
WORKDIR /go/src/github.com/strolt/strolt/apps/stroltm
COPY ./apps/stroltm/go.mod .
COPY ./apps/stroltm/go.sum .
COPY ./shared /go/src/github.com/strolt/strolt/shared
RUN go mod download
COPY ./apps/stroltm/main.go .
COPY ./apps/stroltm/cmd ./cmd
COPY ./apps/stroltm/internal ./internal
COPY --from=ui-builder /github.com/strolt/strolt/apps/stroltm/ui/dist ./internal/ui/build/dist
COPY --from=swagger /github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated /go/src/github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated
RUN go build -o stroltm -ldflags "-X github.com/strolt/strolt/apps/stroltm/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
      org.opencontainers.image.description="Strolt manager" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
      org.opencontainers.image.title="Stroltm" \
      org.opencontainers.image.revision="${GITHUB_SHA}"

COPY --from=builder /go/src/github.com/strolt/strolt/apps/stroltm/stroltm ./bin/stroltm

ENTRYPOINT ["/stroltm/bin/stroltm", "--config", "/stroltm/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/api/v1/ping || exit 1

CMD ["--json"]
