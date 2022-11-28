---
sidebar_position: 1
---

# Events

```yaml
definitions:
  notifications:
    {notification name}:
      driver: {driver name}
      config:
      events:
        - OPERATION_START
        - OPERATION_STOP
        - OPERATION_ERROR

        - SOURCE_START
        - SOURCE_STOP
        - SOURCE_ERROR

        - DESTINATION_START
        - DESTINATION_STOP
        - DESTINATION_ERROR
```


`DEFAULT: OPERATION_STOP, OPERATION_ERROR`
