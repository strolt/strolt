FROM alpine:3.18.0 AS base
WORKDIR /stroltp
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

FROM golang:1.20.4-alpine3.18  AS swagger-base
RUN apk --update add --no-cache make curl

FROM swagger-base AS swagger
WORKDIR /github.com/strolt/strolt
COPY . .
RUN make .swagger-strolt && make .swagger-shared-generate-client-strolt

FROM golang:1.20.4-alpine3.18 AS builder
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

ENTRYPOINT ["/stroltp/bin/stroltp", "--config", "/stroltp/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/api/v1/ping || exit 1

CMD ["start", "--json"]
