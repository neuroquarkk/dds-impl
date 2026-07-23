# dds-impl

Implementations of patterns and concepts from DDS

## Projects

1. [`envsync`](./envsync) *(Sidecar)*
  - Config manager that polls Postgres for configuration changes, updates a local `.env file` and signals the application to reload via `SIGHUP`

2. [`traffic-shadow`](./traffic-shadow) *(Ambassador)*
   - Proxy that mirrors every request to a shadow backend for observation while the client only ever sees the primary's response
