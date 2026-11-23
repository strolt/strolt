# droast global ignore=DF020 reason="manager runs as root by design" expires=2026-12-31
FROM alpine:3.23.7@sha256:ff7f0c727dd796aa36787bfcd9a64bf35a6ce72a73451b3359fde849b6703d63 AS base
WORKDIR /stroltm
# droast ignore=DF052 reason="versions follow the pinned alpine base image" expires=2026-12-31
# hadolint ignore=DL3018
RUN apk --update add --no-cache curl tzdata
ENV TZ=UTC

FROM scratch AS binary
ARG TARGETARCH
COPY dist/stroltm-linux-${TARGETARCH}/stroltm /stroltm

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

COPY --from=binary /stroltm ./bin/stroltm

EXPOSE 8080

ENTRYPOINT ["/stroltm/bin/stroltm", "--config", "/stroltm/config.yml"]

HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "--fail", "http://localhost:8080/api/v1/ping"]

CMD ["--json"]
