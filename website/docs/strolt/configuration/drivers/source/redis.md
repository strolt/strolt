---
sidebar_position: 5
---

# redis

Backs up a self-hosted Redis or Valkey instance by streaming an RDB snapshot
with `redis-cli --rdb`.

```yaml
driver: redis
config:
  host: redis
  port: 6379
  username: default
  password: secret
  # Optional, shell-like extra options for redis-cli; applied to every run.
  # TLS example:
  # common_params: "--tls"
  common_params: ""
  # Optional, extra options applied to backups only.
  backup_params: ""
  # Optional, path to the redis-cli binary (default /usr/bin/redis-cli).
  bin_path_redis_cli: /usr/bin/redis-cli
```

Instead of `host`/`port` a full URI may be given (`uri: redis://redis:6379`).
Any URI userinfo is stripped at config load and never reaches the command
line: the password (`user:pass`, or a colon-less `pass`, which redis-cli
also reads as the password) is delivered through `REDISCLI_AUTH`, the
username through `--user`. Explicit `password`/`username` fields win if both
are set.

## Valkey

Valkey speaks the same protocol and CLI: use this driver and point `host` (or
`uri`) at the Valkey instance.

## Password handling

The password is never passed on the command line (where it would be visible
in `ps`); strolt delivers it to redis-cli through the `REDISCLI_AUTH`
environment variable, which redis-cli reads natively. This holds for the
`password` field and for passwords embedded in `uri`. Passing a secret
explicitly through `common_params`/`backup_params` (for example `-a …`) is
still your own argv — avoid it.

## Pipe mode

`operationMode: pipe` is supported: the RDB stream is handed to the
destination directly (`redis-cli --rdb -`), no intermediate file on disk.

## Restore

Not supported. An RDB file can only be loaded by placing it in the server's
data directory and restarting the instance — there is no way to import it
over the Redis protocol, and strolt does not manage server data directories.
