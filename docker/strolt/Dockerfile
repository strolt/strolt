FROM alpine:3.23.4 AS fetch
ARG TARGETARCH
ARG RESTIC_VERSION=0.19.0

RUN apk add --no-cache \
		postgresql18-client=18.4-r0 \
		mongodb-tools=100.13.0-r6 \
		mariadb-client=11.4.12-r0

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

FROM alpine:3.23.4 AS base
ENV TZ=UTC
WORKDIR /strolt

COPY --from=fetch /apk-lock.txt /apk-lock.txt

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

# MariaDB
COPY --from=fetch /usr/bin/mysqldump  /usr/bin/
COPY --from=fetch /usr/bin/mariadb    /usr/bin/
RUN ln -s /usr/bin/mysqldump /usr/bin/mariadb-dump

# MongoDB tools
COPY --from=fetch /usr/bin/mongodump     /usr/bin/
COPY --from=fetch /usr/bin/mongorestore  /usr/bin/

# Restic
COPY --from=fetch /usr/bin/restic     /usr/bin/

RUN \
	which psql && \
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


FROM golang:1.26.4-alpine3.23 AS builder
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
