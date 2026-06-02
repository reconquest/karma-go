# karma-go

[![Go Reference](https://pkg.go.dev/badge/github.com/reconquest/karma-go.svg)](https://pkg.go.dev/github.com/reconquest/karma-go)

`karma-go` is a Go library for error messages that should be read by
humans under pressure. It keeps the normal Go `error` interface, but renders
wrapping, context, and multi-errors as a hierarchy instead of one long colon
chain.

The standard `errors` package is good at identity: `errors.Is`, `errors.As`,
and unwrapping. It is not trying to be an incident report. `karma-go` fills
that gap without making you give up those checks. Very generous of software to
need two jobs before breakfast.

## Installation

```bash
go get github.com/reconquest/karma-go
```

## Quick start

```go
package main

import (
    "errors"
    "fmt"

    "github.com/reconquest/karma-go"
)

func main() {
    err := karma.Format(
        errors.New("connection refused"),
        "connect to database",
    )

    fmt.Println(err)
}
```

Output:

```text
connect to database
└─ connection refused
```

Each layer says what it was trying to do. The nested reason says why it did
not happen. That is the whole trick: stop hiding structure inside punctuation.

## Why not just `errors`?

Use the standard library for standard-library things. Keep sentinel errors,
typed errors, `errors.Is`, and `errors.As`. Use `karma-go` when the error also
has to explain itself in logs, terminals, alerts, or JSON.

A usual `fmt.Errorf` chain turns a failure path into a sentence:

```go
return fmt.Errorf(
    "handle get user request %q for user %d: scan user row: %w",
    requestID,
    userID,
    err,
)
```

That preserves the wrapped error. It does not preserve structure. The request
ID and user ID are now prose. Your log query has to scrape a sentence. The next
person adds another `: %w`, then another value, and eventually the error line
looks like a municipal tax form.

With `karma-go`, each layer owns the context it knows:

```go
func scanUser(userID int) error {
    err := errors.New("sql: no rows in result set")
    return karma.
        Describe("user_id", userID).
        Format(err, "scan user row")
}

func handleGetUser(requestID string, userID int) error {
    err := scanUser(userID)
    if err != nil {
        return karma.
            Describe("request_id", requestID).
            Format(err, "handle get user request")
    }

    return nil
}
```

Output:

```text
handle get user request
├─ scan user row
│  ├─ sql: no rows in result set
│  └─ user_id: 42
│
└─ request_id: req-7
```

That shape is easier to scan because the hierarchy is visible. It is also
harder to damage by accident: adding a field does not rewrite the whole error
sentence.

## What karma-go gives you

- Hierarchical wrapping with `karma.Format`.
- Key-value context with `karma.Describe`.
- Context-only enrichment with `Reason`.
- Multi-error collection with `karma.Collect`.
- JSON marshaling for tools and APIs.
- Flat output for old loggers with `karma.Flatten`.
- Compatibility with `errors.Is` and `errors.As`.
- Custom text formatting when Unicode is not welcome in your terminal museum.

## Message style

Write messages as lowercase action phrases:

```go
return karma.Format(err, "connect to postgresql")
return karma.Format(err, "load config")
return karma.Format(err, "handle request")
```

Avoid redundant failure words and avoid formatting the nested error into the
message:

```go
// Avoid this.
return fmt.Errorf("failed to connect to postgresql: %w", err)

// Prefer this.
return karma.Format(err, "connect to postgresql")
```

The error level, caller, or returned `error` already says that something failed.
Repeating `failed to` rarely adds useful information.

## Add context where it becomes known

Put fields close to the code that knows them:

```go
func openConfig(path string) error {
    err := errors.New("permission denied")
    if err != nil {
        return karma.
            Describe("path", path).
            Format(err, "open config")
    }

    return nil
}
```

Output:

```text
open config
├─ permission denied
└─ path: /etc/app/config.yml
```

The caller does not need to reconstruct the path from a message string. It is a
field, so loggers and humans can treat it like one.

## Add context without adding another message

Sometimes the error message is already good. Add fields with `Reason`:

```go
func readConfig(path string) error {
    err := errors.New("permission denied")
    if err != nil {
        return karma.
            Describe("path", path).
            Reason(err)
    }

    return nil
}
```

Output:

```text
permission denied
└─ path: /etc/app/config.yml
```

Use `Reason` when you want more diagnostic data, not another narrative layer.

## Wrap at call boundaries

Each function should add the context it owns and then return the error. Higher
layers should not guess lower-layer details.

```go
func parseConfig(path string) error {
    err := errors.New("yaml: line 12: did not find expected key")
    return karma.
        Describe("path", path).
        Format(err, "parse config")
}

func loadConfig(path string) error {
    err := parseConfig(path)
    if err != nil {
        return karma.Format(err, "load config")
    }

    return nil
}

func start() error {
    err := loadConfig("/etc/app/config.yml")
    if err != nil {
        return karma.Format(err, "start service")
    }

    return nil
}
```

Output:

```text
start service
└─ load config
   └─ parse config
      ├─ yaml: line 12: did not find expected key
      └─ path: /etc/app/config.yml
```

A flat version can say the same words. It cannot show ownership as clearly:

```text
start service: load config: parse config: yaml: line 12: did not find expected key | path=/etc/app/config.yml
```

That flat form is still available with `karma.Flatten`. Use it when an old
logger or metrics label cannot handle multiline output. Do not start there if a
human is supposed to read the error.

## Preserve `errors.Is`

`karma-go` keeps the original error in the chain:

```go
var errNotReady = errors.New("not ready")

func waitForService() error {
    return karma.
        Describe("service", "billing").
        Format(errNotReady, "wait for service")
}

func main() {
    err := waitForService()

    fmt.Println(errors.Is(err, errNotReady))
    fmt.Println(err)
}
```

Output:

```text
true
wait for service
├─ not ready
└─ service: billing
```

You get machine-checkable identity and a readable diagnosis in the same value.

## Preserve `errors.As`

Typed errors also survive wrapping:

```go
type RetryableError struct {
    After time.Duration
}

func (err *RetryableError) Error() string {
    return fmt.Sprintf("retry after %s", err.After)
}

func callRemote() error {
    err := &RetryableError{After: 3 * time.Second}
    return karma.
        Describe("endpoint", "https://api.example.com/v1/users").
        Format(err, "call remote api")
}

func main() {
    err := callRemote()

    var retryable *RetryableError
    if errors.As(err, &retryable) {
        fmt.Println("retry delay:", retryable.After)
    }
}
```

`karma-go` is not a replacement for typed errors. It is the wrapper that lets
them explain where they were found.

## Collect several errors

Validation, shutdown, fan-out, and cleanup code often produce more than one
failure. `karma.Collect` keeps them together:

```go
func validateConfig(config map[string]string) error {
    var errs []error

    if config["listen"] == "" {
        errs = append(errs, errors.New("listen address is required"))
    }

    if config["database"] == "" {
        errs = append(errs, errors.New("database url is required"))
    }

    if len(errs) > 0 {
        return karma.Collect("validate config", errs...)
    }

    return nil
}
```

Output:

```text
validate config
├─ listen address is required
└─ database url is required
```

The same value still works with `errors.Is` when the collected branches contain
sentinel errors.

## Add context to collected errors

Each branch can carry its own fields:

```go
func closeShard(name string) error {
    return karma.
        Describe("shard", name).
        Reason(errors.New("connection already closed"))
}

func closeAll() error {
    return karma.Collect(
        "close shards",
        closeShard("eu-west"),
        closeShard("us-east"),
    )
}
```

Output:

```text
close shards
├─ connection already closed
│  └─ shard: eu-west
│
└─ connection already closed
   └─ shard: us-east
```

A single `errors.Join` value can tell you that several things failed. A
`karma-go` tree can show which thing failed where, with the fields still
attached.

## Flatten for old loggers

Some log sinks want one line. Give them one line at the edge:

```go
err := karma.
    Describe("attempt", 3).
    Describe("host", "db.internal").
    Format(errors.New("connection timeout"), "connect to database")

fmt.Println(karma.Flatten(err))
```

Output:

```text
connect to database: connection timeout | attempt=3 host=db.internal
```

Flattening is an output choice, not an error construction strategy. Build the
rich error first. Throw information away only when the sink insists.

## Marshal to JSON

`Karma` and `Context` implement JSON marshaling:

```go
err := karma.
    Describe("service", "auth").
    Describe("user_id", 42).
    Format(errors.New("token expired"), "authenticate request")

data, _ := json.MarshalIndent(err, "", "  ")
fmt.Println(string(data))
```

Output:

```json
{
  "reason": "token expired",
  "message": "authenticate request",
  "context": [
    {
      "key": "service",
      "value": "auth"
    },
    {
      "key": "user_id",
      "value": 42
    }
  ]
}
```

This is useful for error payloads, diagnostic artifacts, or loggers that want
structured data instead of decorative punctuation.

## Customize context value formatting

By default, strings stay strings, scalar values print plainly, and other values
are JSON-formatted. Override `ContextValueFormatter` when a type needs a better
representation:

```go
karma.ContextValueFormatter = func(value interface{}) string {
    switch value := value.(type) {
    case time.Time:
        return value.Format(time.RFC3339)
    case []byte:
        return fmt.Sprintf("<%d bytes>", len(value))
    default:
        return fmt.Sprintf("%v", value)
    }
}
```

Be careful with secrets. `karma-go` will print whatever context you attach,
including tokens, passwords, and regrettable business logic. The library is
helpful, not clairvoyant.

## Customize tree formatting

Text output uses Unicode box drawing by default:

```text
load config
└─ parse config
   └─ permission denied
```

Switch to ASCII delimiters if your log sink or terminal cannot handle Unicode:

```go
karma.BranchDelimiter = karma.BranchDelimiterASCII
karma.BranchChainer = karma.BranchChainerASCII
karma.BranchSplitter = karma.BranchSplitterASCII
karma.BranchIndent = 3
```

Example ASCII output:

```text
load config
\_ parse config
   \_ permission denied
```

## Walk the hierarchy

Use `Descend` when a custom logger wants to process each nested reason:

```go
func writeReason(reason karma.Reason) {
    switch reason := reason.(type) {
    case karma.Karma:
        fmt.Println("message:", reason.GetMessage())
        for _, item := range reason.GetContext().GetKeyValues() {
            fmt.Printf("field: %s=%v\n", item.Key, item.Value)
        }
    default:
        fmt.Println("reason:", reason)
    }
}

func logError(err error) {
    item, ok := err.(karma.Karma)
    if !ok {
        fmt.Println(err)
        return
    }

    writeReason(item)
    item.Descend(writeReason)
}
```

Most applications can just print the error or pass it to their logger. Use
`Descend` when you are building log integration, alerting, or another tool that
cares about the tree.

## Logging pattern

When a logger accepts an error, pass the karma error as the error. The example
uses the `Errorx(error, message, ...args)` call shape; adapt the name to your
logger, not the idea:

```go
err := loadConfig(path)
if err != nil {
    log.Errorx(err, "start service")
    return
}
```

Prefer messages like `start service`, `load config`, and `connect to database`.
Avoid `failed to start service: %v`. If you already have a structured error,
turning it back into a string is vandalism with a format verb.

## Migration from `fmt.Errorf`

A simple migration rule works well:

1. Keep sentinel and typed errors as they are.
2. Replace `fmt.Errorf("verb object: %w", err)` with
   `karma.Format(err, "verb object")`.
3. Move values from the message into `karma.Describe` fields.
4. Add fields at the layer that knows them.
5. Keep using `errors.Is` and `errors.As` for control flow. The wrapper should
   improve diagnostics, not become your new business protocol.

Before:

```go
return fmt.Errorf(
    "sync tenant %q from offset %d: fetch page: %w",
    tenant,
    offset,
    err,
)
```

After:

```go
return karma.
    Describe("tenant", tenant).
    Describe("offset", offset).
    Format(err, "fetch page")
```

Then the caller adds its own layer:

```go
err := fetchPage(tenant, offset)
if err != nil {
    return karma.Format(err, "sync tenant")
}
```

Output:

```text
sync tenant
└─ fetch page
   ├─ connection timeout
   ├─ tenant: acme
   └─ offset: 1200
```

The action chain is readable. The fields are fields. Nobody has to parse a
hand-written sentence while the pager is making its little friendship noises.

## When not to use it

Do not wrap every error just because the function has a return value. A sentinel
or typed error can be enough when the caller only needs control flow.

Do use `karma-go` when one of these is true:

- the error will be logged for an operator or developer;
- a value like path, host, tenant, user ID, request ID, or offset matters;
- several layers contribute useful context;
- several errors must be reported together;
- a tool should serialize or inspect the failure.

The point is not maximal wrapping. The point is useful wrapping.

## API quick reference

```go
karma.Format(reason, "message %s", arg)
karma.Describe("key", value).Format(reason, "message")
karma.Describe("key", value).Reason(reason)
karma.Collect("message", err1, err2, err3)
karma.Flatten(err)
karma.Contains(err, os.ErrNotExist)

var typed *MyError
karma.Find(err, &typed)
```

`Karma` implements `error`, has `Error()` and `String()`, and exposes
`GetMessage`, `GetReasons`, `GetContext`, `Descend`, `Unwrap`, and `Is` for
code that needs to inspect it.

## License

This project is licensed under the terms of the MIT license.
