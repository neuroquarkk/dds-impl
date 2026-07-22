# dds-impl

Implementations of patterns and concepts from DDS

## Projects

1. [`envsync`](./envsync) *(Sidecar)*
  - Config manager that polls Postgres for configuration changes, updates a local `.env file, and signals the application to reload via `SIGHUP`
