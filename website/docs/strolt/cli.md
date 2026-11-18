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

### version

`version` prints only the strolt build version unless `--config <path>` is explicitly supplied. It does not read the default configuration file. With `--config`, it resolves the configured inventory and probes its binaries without running operational configuration validation. Healthy binaries are still listed when another binary in the same driver fails, as are healthy binaries from other drivers. Configuration and probe failures produce a nonzero exit status. Cancellation stops further probes immediately.

Backup task creation requires successful source binary version probes. A missing executable, failed probe, malformed version output, or cancellation aborts task creation rather than adding an error message as a version tag.

## Cancellation

SIGINT or SIGTERM cancels CLI operations, including an idle interactive selection or confirmation. Prompt cleanup restores terminal state before returning; restoration failures are reported alongside cancellation rather than discarded. Cancellation never counts as confirmation. A normal negative confirmation still exits without starting the operation, and `--y` still skips confirmation.

Interactive forms use Huh. Use arrow keys and Enter or Alt+Enter to choose an item. Confirmation accepts `y` or `Y` followed by Enter or Alt+Enter; an empty answer or any other answer declines. End-of-input and Ctrl-C report an error rather than accepting a confirmation.

Newlines inside bracketed paste do not submit a form. Complete the paste and press Enter separately to submit the selection or confirmation.

Failed and canceled CLI task operations remove their temporary work directory before reporting a nonzero exit status. Copy-mode backup errors retain cancellation and subprocess status information as well as the diagnostic text.

Prompts exclusively own standard input while active and release their input reader before another prompt starts. Redirected responses accept LF, CR, and CRLF line endings without consuming the next prompt's answer. Cancellation of an already-issued filesystem or device read remains subject to operating-system and driver completion.

Synchronous API snapshots, stats, and prune operations stop when their HTTP request or daemon lifetime ends. A successfully admitted asynchronous API backup remains owned by the daemon after its HTTP response completes.

Each non-local task uses its own temporary work directory, so rejected concurrent requests cannot remove an active backup's files. Startup does not delete the shared temporary tree. After a crash or forced termination, leftovers may remain under `strolt-data/tmp` in the configured data directory; stop all strolt processes sharing that directory before cleaning those leftovers manually.
