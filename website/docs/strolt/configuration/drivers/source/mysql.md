---
sidebar_position: 3
---

# mysql

```yaml
driver: mysql
config:
  host: mysql
  port: 3306
  database: app
  username: app
  password: secret
  # Optional. TLS behavior of the bundled MariaDB client:
  #   (empty)     - client defaults (verifies the server certificate)
  #   skip-verify - use TLS but do not verify the server certificate;
  #                 required for MySQL 8+ servers with auto-generated
  #                 self-signed certificates
  #   disabled    - do not use TLS
  tls: skip-verify
```
