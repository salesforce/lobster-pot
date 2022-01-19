# Adding a new scanner engine

You can add a new scanner engine either as a binary or as a golang vendored scanner.

## Golang library

In case of using a Golang library, it needs to be added as an import, and vendored with the project.

1. Add a new go file in the [/scanner/](../../scanner/) directory, with at least a function returning a `Findings` struct
1. Add a case in the `switch` statement in [scanner/embedded.go](../../scanner/embedded.go) for this new scanner.
1. Add an entry in the [config/scanner.go](../../config/scanner.go) file, in the `IsValidScannerName`function.
1. Update the [docs/configurations/scanner.md](../configuration/scanner.md) file to include the new parser.

The [scanner/wraith.go](../../scanner/wraith.go) file can be taken as an example.

## External binary

To be able to add an external binary, multiple solutions are possible:

- Add a `bin/go-pre-compile` file containing a bash script that downloads the binary when deploying the application.
- Leverage the various buildpacks available to install the binary (see [Semgrep](semgrep.md) for an example).

### Parser

1. Add a new parser in the `scanner` package. The function signature is `func Parse<NewParser>Findings(data []byte) (Findings, error)`
1. Update the `parseFindingsFromBinaryScanner` function in the [scanner/binary.go](../../scanner/binary.go) file to include the new parser.
1. Add an entry in the [config/scanner.go](../../config/scanner.go) file, in the `IsValidScannerName`function.
1. Update the [docs/configuarations/scanner.md](../configuration/scanner.md) file to include the new parser.
1. Add a new document in the [docs/scanners/](.) directory, including configuration vars examples.

The [scanner/semgrep.go](../../scanner/semgrep.go) file can be taken as an example.

The binary needs to be locally available in the app's slug. If deploying to Heroku, or similar environment, it is possible to run a build script to download binaries using the `bin/go-pre-compile` script.
