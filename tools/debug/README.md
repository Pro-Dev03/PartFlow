# Development Debug Utilities

This folder contains temporary debugging and probe scripts used during local development.

## Important
- These files are for development and troubleshooting only.
- They are intentionally kept outside the main production runtime paths.
- Do not treat them as part of the shipped product.
- Do not use hardcoded machine-specific paths in production code.

## Typical usage
- DB probes
- login verification helpers
- owner seeding helpers
- environment validation tools

## Production rule
Production code must resolve database paths dynamically using:
- PARTFLOW_LOCAL_DB_PATH environment variable, or
- os.UserConfigDir()/application data directory, or
- explicit app configuration
