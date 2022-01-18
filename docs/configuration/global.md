# Global configuration variables

## Environment

`ENVIRON`: The current environment.
It can be set to `dev` to be able to enable the `trace` logging level, and to bypass the `GITHUB_SECRET` check, for local development.

## Database

`DATABASE_URL`: The URL of the postgres database to use. The format is `postgres://username:password@host:port/db_name`

## Logging and error reporting

`LOG_LEVEL`: The level of logging to use. Defaults to `info`. Can be one of `trace`, `debug`, `info`, `warn`, `error`, `fatal`.
The `trace` level can only be activated in the `dev` environment.

`ROLLBAR_TOKEN` - The token to use for reporting errors to Rollbar