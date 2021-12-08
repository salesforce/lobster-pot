# Using Semgrep as a scanner for secret scanning

## Install binary

- add a `requirements.txt` file with content `semgrep==0.68.2`
- add a `runtime.txt` file with content `python-3.9.7`

## Install rules

- Add a `bin/go-pre-compile` file containing :

```bash
#!/bin/bash

git clone https://github.com/returntocorp/semgrep-rules.git semgrep-rules
```

We can't add it as a git submodule for the project becauses it messes up with the `go` modules

## Configure environment

```bash
SCANNER_BINARY=semgrep
SCANNER_ARGUMENTS="--config=/app/semgrep-rules/generic/secrets/security/;--exclude=vendor/;--exclude=semgrep-rules/;--exclude=bin/;--json;%s"
SCANNER_NAME=semgrep
```
