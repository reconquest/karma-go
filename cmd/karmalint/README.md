# karmalint

A Go linter that detects and auto-fixes common return error clauses, migrating
code to idiomatic [`karma-go`](https://github.com/reconquest/karma-go) usage.

## Installation

```bash
go install github.com/reconquest/karma-go/cmd/karmalint@latest
```

## Usage

```bash
# Check for issues
karmalint ./...

# Auto-fix issues
karmalint -fix ./...

# Check specific file or directory
karmalint path/to/file.go
karmalint path/to/directory
```

## What It Detects

### 1. `fmt.Errorf` Error Wrapping

Converts `fmt.Errorf("message: %w", err)` to `karma.Format(err, "message")`
for hierarchical error wrapping. The `karma` import is added automatically
when missing.

```go
// Bad
return fmt.Errorf("failed to connect: %w", err)

// Good (auto-fixed)
return karma.Format(err, "connect")
```

Also handles multiple format arguments, including complex expressions:

```go
// Bad
return fmt.Errorf("failed to process %s: %w", cfg.Name, err)

// Good (auto-fixed)
return karma.Format(err, "process %s", cfg.Name)
```

A trailing `: %v` / `: %s` with an error argument is flagged as well, because
formatting the error into the message destroys the error chain
(`errors.Is`/`errors.As` stop working):

```go
// Bad
return fmt.Errorf("failed to connect: %v", err)

// Good (auto-fixed)
return karma.Format(err, "connect")
```

The last argument is only treated as an error when its name looks like one
(`err`, `error`, `parseError`, `customErr`, ...), and the number of format
verbs MUST match the number of arguments. `fmt.Errorf` calls inside the
`karma` package itself and files that import karma-go under a custom name
are left alone.

### 2. Redundant Prefixes in Error Messages

Phrases like "failed to", "unable to", "could not" are redundant in
`karma.Format` and `karma.Collect` messages: the returned error already
indicates failure, and the hierarchy shows where it happened.

```go
// Bad
return karma.Format(err, "failed to connect to database")
return karma.Describe("path", path).Format(err, "unable to open config")

// Good (auto-fixed)
return karma.Format(err, "connect to database")
return karma.Describe("path", path).Format(err, "open config")
```

**Detected patterns:**
- `failed to ...`
- `unable to ...`
- `could not ...`
- `cannot ...`
- `can't ...`
- `error ...`
- `err: ...`

### 3. Trailing Failure Words

Removes redundant trailing words like "failed", "error", "err" from error
messages.

```go
// Bad
return karma.Format(err, "database connection failed")

// Good (auto-fixed)
return karma.Format(err, "database connection")
```

## Supported Calls

| Call | Message position |
|------|------------------|
| `karma.Format(reason, "message", args...)` | argument 1 |
| `karma.Describe(...).Format(reason, "message", ...)` | argument 1 |
| `karma.Collect("message", errs...)` | argument 0 |
| `fmt.Errorf("message: %w", ..., err)` | converted to `karma.Format` |

## Notes

- `-fix` runs changed files through goimports when they were gofmt-clean:
  imports left unused by the conversions (for example `fmt`) are removed and
  grouping is normalized. Files with other formatting keep their layout.
- Only string literal messages are checked; dynamically built messages are
  skipped.
- Vendor and hidden directories are skipped during recursive scans.

## Exit Codes

- `0`: No issues found (or all issues fixed with `-fix`)
- `1`: Issues found

## License

MIT
