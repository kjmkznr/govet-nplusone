# govet-nplusone

Static analysis tool (go/analysis) to detect potential SQL N+1 query patterns. As a first step, it targets `database/sql` and reports calls to `Query*`/`Exec*`/`Prepare*` that happen inside loops.

For the Japanese version of this document, see [README.ja.md](README.ja.md)

## Features
- Detects `database/sql` method calls that occur inside loops like `for`/`range`
- Target methods: `Query`, `QueryContext`, `QueryRow`, `QueryRowContext`, `Exec`, `ExecContext`, `Prepare`, `PrepareContext`
- Supports standalone execution via `singlechecker`; easy to integrate with `go vet -vettool`

## Requirements
- Go Modules
- Go 1.20+ recommended (depends on `golang.org/x/tools` requirements)

## Build

```shell
go build ./...
```

## Install

```shell
go install ./cmd/nplusone
```

This produces the `nplusone` command under `$GOPATH/bin` (or `GOBIN`).

## Usage
### Run as a standalone analyzer
```
# Analyze current directory recursively
nplusone ./...

# Analyze a specific package path
nplusone ./pkg/...
```

### Integrate with go vet (-vettool)
```
# Ensure nplusone is installed first
VETTOOL=$(which nplusone)

go vet -vettool="$VETTOOL" ./...
```

## Example output
Given the following kind of process:

```go
for _, id := range ids {
    _ = id
    _ = db.QueryRowContext(ctx, "SELECT 1") // want "potential N\\+1: database/sql method QueryRowContext called inside a loop"
}
```

You may see a report like:

```
path/to/file.go:NN:NN: potential N+1: database/sql method QueryRowContext called inside a loop
```

## Detection logic (overview)
- Traverse the AST using the `inspect` pass and keep a depth count for entering `ForStmt`/`RangeStmt`
- Only consider `CallExpr` inside loops; determine whether the method belongs to `database/sql` via type info (`pass.TypesInfo.Selections`)
- Exclude package-level functions (e.g., `sql.Open`) which don't have selection info

## Limitations / Known issues
- Currently only targets `database/sql`. ORMs (gorm, sqlx, etc.) and wrapper functions are not supported yet
- Emits a warning about a potential N+1 rather than a definitive finding (keeps rules simple to reduce false positives)
- Does not yet detect or suppress cases optimized outside the loop (prepared statement reuse, batching, etc.)
- No suppression comments or config-based exclusions yet

## License
- MIT
