---
sidebar_position: 2
---

# Configuration

## Status of this document

This document specifies the Compose file format used to define multi-containers applications. Distribution of this document is unlimited.

The key words `MUST`, `MUST NOT`, `REQUIRED`, `SHALL`, `SHALL NOT`, `SHOULD`, `SHOULD NOT`, `RECOMMENDED`, `MAY`, and `OPTIONAL` in this document are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119).

## Strolt file

The Strolt file is a YAML file defining services (REQUIRED), timezone, tags, secrets, extends and definitions. The default path for a Strolt file is strolt.yaml (preferred) or strolt.yml in working directory. If both files exist, Strolt implementations MUST prefer canonical strotl.yaml one.

```yaml
# IANA Time Zone Default: Local
timezone: UTC

disableWatchChanges: false

tags:
  - tag1:value
  - tag2:value
  - tag3:value

services:
  service-name:
    task-name:
      schedule:
        backup: "*/5 * * * *"
        prune: "*/5 * * * *"
      tags:
        - tag1:value
        - tag2:value
        - tag3:value
      source:
      destinations:
      notifications:
        - telegram

secrets:
  SUPER_SECRET: SECRET

extends:
  secrets:
    - ./secrets.yml
  configs:
    - ./strolt_config.yml

definitions:
  destinations:
    destination-template:
      driver: restic
      config:
        keep:
          last: 5

  notifications:
    telegram:
      driver: telegram
      config:
        token: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
        chatId: 1234567890
```

## TimeZone top-level element

`timezone`: IANA Time Zone Default: Local

```yaml
timezone: UTC
```

## Tags top-level element

`tags`: array of strings

```yaml
tags:
  - tag1:v0.01
  - tag2
```

**OR**

```yaml
tags: [tag1:v0.01, tag2]
```

## Services top-level element `(REQUIRED)`

`services`: map of services

```yaml
services:
  service-name:
    task-name:
      tags:
      schedule:
      source:
        driver: pg
          config:
            host: 127.0.0.1
            port: 5432
            database: pg_database
            username: pg_username
            password: pg_password
      destinations:
      notifications:

```

### tags

`tags`: array of strings

```yaml
tags:
  - tag1:v0.01
  - tag2
```

**OR**

```yaml
tags: [tag1:v0.01, tag2]
```

### schedule

`schedule`:

```yaml
schedule:
  # run every day at 6:00 and 18:00 UTC
  backup: "0 6,18 */1 * *"

  # run every day at 6:00 and 18:00 UTC
  prune: "0 6,18 */1 * *"
```

#### backup

`backup`: cron rule

#### prune

`prune`: cron rule

### source `(REQUIRED)`

`source`:

```yaml
source:
  driver: pg
    config:
      host: 127.0.0.1
      port: 5432
      database: pg_database
      username: pg_username
      password: pg_password
```

#### driver `(REQUIRED)`

`driver`:

#### config

`config`:

#### env

`env`:

### destinations `(REQUIRED)`

`destinations`:

```yaml
services:
  {service name}:
    {task name}:
      ...
      destinations:
        restic:
          driver: restic
          config:
          env:

```

**OR**

```yaml
services:
  {service name}:
    {task name}:
      ...
      destinations:
        restic:
          extends: {definition template}

definitions:
  destinations:
    {definition template}:
      driver: restic
      config:
      env:

```

#### extends

`extends`:

#### driver `(REQUIRED)`

`driver`: [source](./drivers/source/local.md)

#### config

`config`:

#### env

`env`:

### notifications

`notifications`:

```yaml
services:
  {service name}:
    {task name}:
      ...
      notifications:
        - tg

    {task2 name}:
      ...
      notifications:
        - tg-errors

definitions:
  notifications:
    tg:
      driver: telegram
      config:
        token: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
        chatId: 1234567890
    tg-errors:
      driver: telegram
      config:
        token: 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
        chatId: 1234567890
      events: [OPERATION_ERROR, SOURCE_ERROR, DESTINATION_ERROR]
```

## Secrets top-level element

`secrets`: map of key value

```yaml
secrets:
  SECRET_ONE: VALUE
  SECRET_TWO: VALUE
```

:::info

Store important data in files with secrets that are included in the "extends" section

:::

## Extends top-level element

### secrets

`secrets`: array of paths

```yaml
extends:
  secrets:
    - strolt.secrets.yml
```

### configs

`configs`: array of paths

```yaml
extends:
  configs:
    - strolt.base.yml
```

## Definitions top-level element

### destinations

#### driver `(REQUIRED)`

`driver`: [source](./drivers/source/local.md)

#### config

`config`:

#### env

`env`:

### notifications

#### driver `(REQUIRED)`

`driver`: [source](./drivers/source/local.md)

#### config

`config`:

#### events

`events`: array of [`OPERATION_START`, `OPERATION_STOP`, `OPERATION_ERROR`, `SOURCE_START`, `SOURCE_STOP`, `SOURCE_ERROR`, `DESTINATION_START`, `DESTINATION_STOP`, `DESTINATION_ERROR`]
