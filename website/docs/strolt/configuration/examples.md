---
sidebar_position: 3
---

# Examples

## Example1

```yaml title=strolt.yml

```

## Example2

```yaml title=strolt.yml
timezone: America/New_York

tags:
  - tag1:value
  - tag2:value
  - tag3:value

services:
  example:
    database:
      tags:
        - taskTag
      schedule:
        backup: "0 0 * * *"
        prune: "0 20 * * *"
      source:
        driver: pg
        config:
          host: "{{.PG_HOST}}"
          database: "{{.PG_DATABASE}}"
          username: "{{.PG_USER}}"
          password: "{{.PG_PASSWORD}}"
      destinations:
        restic-postgres:
          extends: minio_restic
          env:
            RESTIC_PASSWORD: "{{.RESTIC_PASSWORD}}"
            RESTIC_REPOSITORY: "s3:{{.MINIO_S3_URL}}/EXAMPLE2/restic-postgres"
      notifications:
        - telegram
        - errors

secrets:
  RESTIC_PASSWORD: password

  PG_HOST: postgres
  PG_DATABASE: strolt
  PG_USER: strolt_user
  PG_PASSWORD: strolt_password

extends:
  secrets:
    - ./strolt.secrets.yml
  configs:
    - ./strolt.base.yml
```

```yaml title=strolt.base.yml
definitions:
  notifications:
    telegram:
      driver: telegram
      config:
        token: "{{.TELEGRAM_TOKEN}}"
        chatId: "{{.TELEGRAM_CHAT_ID}}"
    errors:
      driver: telegram
      config:
        token: "{{.TELEGRAM_TOKEN}}"
        chatId: "{{.TELEGRAM_CHAT_ID}}"
      events:
        - OPERATION_ERROR
        - SOURCE_ERROR
        - DESTINATION_ERROR

  destinations:
    minio_restic:
      driver: restic
      config:
        keep:
          last: 3
      env:
        AWS_ACCESS_KEY_ID: "{{.MINIO_S3_ACCESS_KEY_ID}}"
        AWS_SECRET_ACCESS_KEY: "{{.MINIO_S3_SECRET_ACCESS_KEY}}"
```

```yaml title=strolt.secrets.yml
MINIO_S3_URL: http://minio:9000
MINIO_S3_ACCESS_KEY_ID: minioadmin
MINIO_S3_SECRET_ACCESS_KEY: minioadmin

TELEGRAM_TOKEN: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_CHAT_ID: 1234567890
```

```yaml title=docker-compose.yml
services:
  minio:
    image: minio/minio:RELEASE.2022-08-13T21-54-44Z
    command: server --address 0.0.0.0:9000 --console-address 0.0.0.0:9001 /data
    ports:
      - 9000:9000
      - 9001:9001
    volumes:
      - minio:/data
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    networks:
      - strolt_network

  postgres:
    image: postgres:13.2-alpine
    environment:
      TZ: UTC
      POSTGRES_DB: strolt
      POSTGRES_PASSWORD: strolt_password
      POSTGRES_USER: strolt_user
    volumes:
      - postgres:/var/lib/postgresql/data
    networks:
      - strolt_network

  strolt:
    image: strolt/strolt:latest
    depends_on:
      - postgres
    restart: always
    environment:
      - STROLT_GLOBAL_TAGS=postgres:13.2-alpine
    volumes:
      - ./strolt.yml:/app/strolt/config.yml:ro
      - ./strolt.secrets.yml:/app/strolt/strolt.secrets.yml:ro
      - ./strolt.base.yml:/app/strolt/strolt.base.yml:ro
    networks:
      - strolt_network

networks:
  strolt_network:

volumes:
  minio:
  postgres:
```
