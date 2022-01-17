# Scanner configuration

## Golang vendored scanner

In order to use a golang vendored scanner, you need to use the following configuration variables:

```bash
SCANNER_TYPE=golang
SCANNER_NAME=<chosen scanner>
```
### Binary scanner

in the [docs/scanners](../scanners) directory you can find a list of available scanners, with their description and suggested configuration variables.

The relevant variables are:

```bash
SCANNER_TYPE=binary
SCANNER_NAME=<chosen scanner>
SCANNER_BINARY=<binary>
SCANNER_ARGUMENTS="arguments;split;by;semi-colon;with;%s;as;placeholder;for;path"
```

The binary needs to be locally available in the app's slug. If deploying to Heroku, or similar environment, it is possible to run a build script to download binaries using the `bin/go-pre-compile` script

To configure a new scanner, see the [docs/scanners](../scanners) directory.