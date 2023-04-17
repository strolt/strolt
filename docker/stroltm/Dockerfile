FROM alpine:3.17.3 AS base
WORKDIR /stroltm
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

FROM node:18.15.0-bullseye-slim as ui-base
RUN apt-get update && apt-get -y install openjdk-11-jre --no-install-recommends && apt-get clean && rm -rf /var/lib/apt/lists/*

FROM --platform=linux/amd64 golang:1.20.3-alpine3.17  AS swagger-base
RUN apk --update add --no-cache make curl

FROM  --platform=linux/amd64 ui-base as ui-deps
WORKDIR /github.com/strolt/strolt/apps/stroltm/ui
COPY ./apps/stroltm/ui/package.json .
COPY ./apps/stroltm/ui/yarn.lock .
RUN yarn install --frozen-lockfile --prefer-offline --ignore-scripts --network-timeout 1000000 && yarn cache clean

FROM swagger-base AS swagger
WORKDIR /github.com/strolt/strolt
COPY . .
RUN make .swagger-strolt && make .swagger-shared-generate-client-strolt && make .swagger-stroltp && make .swagger-shared-generate-client-stroltp && make .swagger-stroltm

FROM ui-deps as ui-builder
COPY ./apps/stroltm/ui .
COPY --from=swagger /github.com/strolt/strolt/.swagger /github.com/strolt/strolt/.swagger
RUN yarn gen-api && yarn build

FROM golang:1.20.3-alpine3.17 AS builder
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
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/strolt/generated /go/src/github.com/strolt/strolt/shared/sdk/strolt/generated
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/stroltp/generated /go/src/github.com/strolt/strolt/shared/sdk/stroltp/generated
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
