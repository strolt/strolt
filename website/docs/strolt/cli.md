---
sidebar_position: 3
---

# CLI

## Commands

### init

`init`:

### backup

`backup`:

### restore

`restore`:

### prune

`prune`:

### forget

`forget`: removes a single snapshot from a destination and prunes its data. The retention policy of the destination is ignored, so a snapshot the `keep` rules would preserve is deleted as well.

```sh
strolt forget --service <service> --task <task> --destination <destination> --snapshot <snapshot id>
```

Every flag is optional; missing values are asked for interactively. `--y` skips the confirmation prompt.

### unlock

`unlock`: removes locks left in the destination repository, for example after a backup process was killed and later operations fail with "repository is already locked".

```sh
strolt unlock --service <service> --task <task> --destination <destination>
```

By default only stale locks are removed — locks whose owning process is gone. Add `--remove-all` to also remove the locks of operations that are still running.

### config

`config`:
