# droast global ignore=DF020 reason="runs as root by design: data dir and dump tooling" expires=2026-12-31
FROM alpine:3.23.7@sha256:ff7f0c727dd796aa36787bfcd9a64bf35a6ce72a73451b3359fde849b6703d63 AS fetch
ARG TARGETARCH
ARG RESTIC_VERSION=0.19.1

RUN apk add --no-cache \
		postgresql18-client=18.7-r0 \
		mongodb-tools=100.13.0-r6 \
		mariadb-client=11.4.14-r0 \
		mariadb-connector-c=3.4.6-r0

# restic is fetched from official releases: alpine still packages 0.18.x
RUN wget -q \
		"https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2" \
		"https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/SHA256SUMS" && \
	grep "restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2" SHA256SUMS > restic.sha256 && \
	sha256sum -c restic.sha256 && \
	bunzip2 "restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2" && \
	install -m 0755 "restic_${RESTIC_VERSION}_linux_${TARGETARCH}" /usr/bin/restic && \
	rm -f SHA256SUMS restic.sha256 "restic_${RESTIC_VERSION}_linux_${TARGETARCH}"

RUN rm -f \
		/usr/bin/clusterdb \
		/usr/bin/createdb \
		/usr/bin/createuser \
		/usr/bin/dropdb \
		/usr/bin/dropuser \
		/usr/bin/pg_basebackup \
		/usr/bin/pg_dumpall \
		/usr/bin/pg_isready \
		/usr/bin/pg_receivewal \
		/usr/bin/pg_recvlogical \
		/usr/bin/reindexdb \
		/usr/bin/vacuumdb \
		/usr/bin/pg_amcheck \
		/usr/bin/pgbench \
		/usr/bin/pg_verifybackup

RUN apk info -v > /apk-lock.txt

FROM alpine:3.23.7@sha256:ff7f0c727dd796aa36787bfcd9a64bf35a6ce72a73451b3359fde849b6703d63 AS base
ENV TZ=UTC
WORKDIR /strolt

COPY --from=fetch /apk-lock.txt /apk-lock.txt

# droast ignore=DF052 reason="runtime libs follow the pinned alpine base; fetch stage pins exact versions" expires=2026-12-31
# hadolint ignore=DL3018
RUN apk add --no-cache \
		curl \
		tzdata\
		libpq \
		libssl3 \
		libcrypto3 \
		ca-certificates \
		readline \
		lz4-libs \
		libstdc++ \
		libgcc \
		krb5-libs
RUN update-ca-certificates

# PostgreSQL
COPY --from=fetch /usr/bin/psql       /usr/bin/
COPY --from=fetch /usr/bin/pg_dump    /usr/bin/
COPY --from=fetch /usr/bin/pg_restore /usr/bin/

# MariaDB. The client auth plugins (caching_sha2_password and friends) are
# required to connect to MySQL 8+ servers.
COPY --from=fetch /usr/bin/mysqldump  /usr/bin/
COPY --from=fetch /usr/bin/mariadb    /usr/bin/
COPY --from=fetch /usr/lib/mariadb/plugin/ /usr/lib/mariadb/plugin/
RUN ln -s /usr/bin/mysqldump /usr/bin/mariadb-dump

# MongoDB tools
COPY --from=fetch /usr/bin/mongodump     /usr/bin/
COPY --from=fetch /usr/bin/mongorestore  /usr/bin/

# Restic
COPY --from=fetch /usr/bin/restic     /usr/bin/

RUN \
	command -v psql && \
	psql --version && \
	pg_dump --version && \
	pg_restore --version && \
	mariadb --version && \
	mysqldump --version && \
	mongodump --version && \
	mongorestore --version && \
	restic version

RUN \
		ldd /usr/bin/psql && \
		ldd /usr/bin/pg_dump && \
		ldd /usr/bin/pg_restore && \
		ldd /usr/bin/mysqldump && \
		ldd /usr/bin/mariadb && \
		ldd /usr/bin/mongodump && \
		ldd /usr/bin/mongorestore
# restic is excluded from the ldd check: official release binaries are static

FROM --platform=$BUILDPLATFORM node:26.13.1-trixie-slim AS swagger
# Datamitsu needs git to identify the workspace and system CA roots for HTTPS downloads.
# droast ignore=DF005 reason="ca-certificates and git are managed by the Debian repository" expires=2026-12-31
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && rm -rf /var/lib/apt/lists/*
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
COPY ./shared ./shared
RUN pnpm dm init --skip-download --quiet && pnpm dm exec task -- codegen:go -- strolt && rm -rf .git


FROM golang:1.27.3-alpine3.23 AS builder-deps
WORKDIR /go/src/github.com/strolt/strolt/apps/strolt
COPY ./apps/strolt/go.mod .
COPY ./apps/strolt/go.sum .
COPY ./shared /go/src/github.com/strolt/strolt/shared
RUN go mod download

FROM builder-deps AS builder
ARG version
COPY ./apps/strolt/main.go .
COPY ./apps/strolt/cmd ./cmd
COPY ./apps/strolt/internal ./internal
COPY --from=swagger /github.com/strolt/strolt/shared/sdk/strolt/generated /go/src/github.com/strolt/strolt/shared/sdk/strolt/generated
RUN go build -o strolt -ldflags "-X github.com/strolt/strolt/apps/strolt/internal/ldflags.version=${version} -s -w"

FROM base
ARG GITHUB_SHA
ARG RELEASE_VERSION=development
ARG RELEASE_RUN_ID=0
ARG RELEASE_ATTEMPT=0
ARG RELEASE_WORKFLOW_SHA=development
ARG RELEASE_CHANNEL=development
ENV STROLT_PATH_DATA=/strolt

LABEL org.opencontainers.image.authors="shibanet0 <shibanet0@gmail.com>" \
			org.opencontainers.image.description="Strolt backup engine" \
			org.opencontainers.image.licenses="MIT" \
			org.opencontainers.image.source="https://github.com/strolt/strolt.git" \
			org.opencontainers.image.title="Strolt" \
            org.opencontainers.image.revision="${GITHUB_SHA}" \
            org.opencontainers.image.version="${RELEASE_VERSION}" \
            com.strolt.release.run="${RELEASE_RUN_ID}" \
            com.strolt.release.attempt="${RELEASE_ATTEMPT}" \
            com.strolt.release.workflow="${RELEASE_WORKFLOW_SHA}" \
            com.strolt.release.channel="${RELEASE_CHANNEL}" \
            com.strolt.release.product="strolt"

COPY --from=builder /go/src/github.com/strolt/strolt/apps/strolt/strolt ./bin/strolt

EXPOSE 8080

ENTRYPOINT ["/strolt/bin/strolt", "--config", "/strolt/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "--fail", "http://localhost:8080/api/v1/ping"]

CMD ["start", "--json"]
