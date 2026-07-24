---
sidebar_position: 1
---

# restic

```yaml
driver: restic
```

## Tuning

All options below are optional. When they are left out, strolt runs restic exactly as it did before they existed.

### Throughput

| Option             | restic flag              | Notes                                                                                                                                          |
| ------------------ | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `read-concurrency` | `--read-concurrency <n>` | Files read in parallel during backup. restic defaults to 2, which is low for FUSE or object-storage sources.                                   |
| `no-scan`          | `--no-scan`              | Skips the upfront tree walk that only estimates the backup size; avoids a LIST/HEAD storm on high-latency sources.                             |
| `pack-size`        | `--pack-size <n>`        | Target pack size in MiB. Larger packs mean fewer objects in the repository.                                                                    |
| `options`          | `-o <key>=<value>`       | restic extended options, one flag per entry. Applies to every restic call, so `rest.connections` speeds up both uploads and restore downloads. |

```yaml
driver: restic
config:
  read-concurrency: 8
  no-scan: true
  pack-size: 64
  options:
    - rest.connections=10
```

### Restore target

`restore-target` restores into the given directory instead of the task work directory:

```yaml
driver: restic
config:
  restore-target: /var/lib/strolt/restore
```

This is meant for sources that cannot serve the concurrent random writes of the restorer, such as a FUSE mountpoint. Note that strolt does not move the restored data into the source afterwards — the snapshot is left in the configured directory, and a warning is logged for every restore that uses it.

### Escape hatches

Flags and environment variables strolt does not model explicitly can be passed through:

```yaml
driver: restic
config:
  extra_backup_args:
    - --exclude-caches
  extra_restore_args:
    - --verify
  env_extra:
    RESTIC_PASSWORD_COMMAND: "cat /run/secrets/restic"
env:
  RESTIC_REPOSITORY: "s3:http://minio:9000/repo"
```

`extra_backup_args` and `extra_restore_args` are appended to the end of `restic backup` and `restic restore`, so they take precedence over the flags strolt builds. `env_extra` is required for any variable the `env` block does not support: strolt builds the restic environment from scratch instead of inheriting the one it was started with, so nothing else reaches restic. Its pairs are applied last and therefore override `env`. It lives under `config` because `env` only accepts flat key/value pairs of known variables.
