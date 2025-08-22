# karma-go

[![Go Reference](https://pkg.go.dev/badge/github.com/reconquest/karma-go.svg)](https://pkg.go.dev/github.com/reconquest/karma-go)

A Go library for creating hierarchical, context-rich error messages with beautiful formatting.

## Overview

karma-go transforms flat error messages into structured, hierarchical representations that are easier to debug and understand. It provides:

- **Hierarchical Error Structure**: Build nested error chains that show the complete failure path
- **Rich Context**: Add key-value pairs to errors for better debugging
- **Beautiful Formatting**: Pretty-printed error trees with Unicode box drawing characters
- **JSON Serialization**: Full support for marshaling/unmarshaling errors
- **Multi-error Support**: Collect and display multiple errors together
- **Fluid Interface**: Chain method calls for clean, readable code

## Installation

```bash
go get github.com/reconquest/karma-go
```

## Quick Start

### Basic Error Formatting

```go
package main

import (
    "fmt"
    "github.com/reconquest/karma-go"
)

func main() {
    err := karma.Format(
        fmt.Errorf("connection refused"),
        "failed to connect to database",
    )
    
    fmt.Println(err)
    // Output:
    // failed to connect to database
    // └─ connection refused
}
```

### Adding Context

```go
func connectDB(host string, port int) error {
    // Simulate connection failure
    err := fmt.Errorf("connection refused")
    
    return karma.
        Describe("host", host).
        Describe("port", port).
        Format(err, "failed to connect to database")
}

func main() {
    err := connectDB("localhost", 5432)
    fmt.Println(err)
    // Output:
    // failed to connect to database
    // ├─ connection refused
    // ├─ host: localhost
    // └─ port: 5432
}
```

### Nested Error Chains

```go
func queryUser(id int) error {
    err := connectDB("localhost", 5432)
    if err != nil {
        return karma.
            Describe("user_id", id).
            Format(err, "failed to query user")
    }
    return nil
}

func handleRequest() error {
    err := queryUser(123)
    if err != nil {
        return karma.Format(err, "request processing failed")
    }
    return nil
}

func main() {
    err := handleRequest()
    fmt.Println(err)
    // Output:
    // request processing failed
    // └─ failed to query user
    //    ├─ failed to connect to database
    //    │  ├─ connection refused
    //    │  ├─ host: localhost
    //    │  └─ port: 5432
    //    │
    //    └─ user_id: 123
}
```

## Key Features

### Context Support

Add arbitrary key-value fields to errors for enhanced debugging:

```go
func processFile(filename string) error {
    err := fmt.Errorf("file not found")
    return karma.
        Describe("filename", filename).
        Describe("operation", "read").
        Describe("timestamp", time.Now()).
        Reason(err)
}
```

### Multi-error Collection

Collect multiple errors into a single hierarchical structure:

```go
func validateInput(data map[string]string) error {
    var errors []error
    
    if data["name"] == "" {
        errors = append(errors, fmt.Errorf("name is required"))
    }
    if data["email"] == "" {
        errors = append(errors, fmt.Errorf("email is required"))
    }
    
    if len(errors) > 0 {
        return karma.Collect("validation failed", errors...)
    }
    return nil
}
```

### JSON Serialization

Full support for JSON marshaling and unmarshaling:

```go
err := karma.
    Describe("service", "auth").
    Format(fmt.Errorf("token expired"), "authentication failed")

data, _ := json.Marshal(err)
fmt.Println(string(data))
// Output: {"reason":"token expired","message":"authentication failed","context":[{"key":"service","value":"auth"}]}
```

### Error Inspection

Find specific error types in the error chain:

```go
var customErr *MyCustomError
if karma.Find(err, &customErr) {
    // Handle specific error type
}

// Check if error chain contains specific error
if karma.Contains(err, os.ErrNotExist) {
    // Handle file not found
}
```

### Flattening for Logging

Convert hierarchical errors to flat strings for traditional logging:

```go
err := buildComplexError()
fmt.Println(karma.Flatten(err))
// Output: request failed: database error: connection timeout | host=localhost port=5432 retry_count=3
```

## Customization

### Custom Formatting

Customize the appearance of error trees:

```go
// Use ASCII characters instead of Unicode
karma.BranchDelimiter = karma.BranchDelimiterASCII  // "\_ "
karma.BranchChainer = karma.BranchChainerASCII      // "| "
karma.BranchSplitter = karma.BranchSplitterASCII    // "+ "

// Adjust indentation
karma.BranchIndent = 4
```

### Custom Context Value Formatting

Control how context values are displayed:

```go
karma.ContextValueFormatter = func(value interface{}) string {
    switch v := value.(type) {
    case time.Time:
        return v.Format(time.RFC3339)
    case []byte:
        return fmt.Sprintf("<%d bytes>", len(v))
    default:
        return fmt.Sprintf("%v", value)
    }
}
```

## Comparison with Other Libraries

| Feature                  | karma-go | pkg/errors | emperror |
|--------------------------|----------|------------|----------|
| Nested Errors            | ✔        | ✔          | ✔        |
| Key-Value Context        | ✔        | ✗          | ~        |
| Descriptive Pretty Print | ✔        | ✗          | ✗        |
| Embedded Stack Trace     | ✗        | ✔          | ✔        |
| JSON Friendly            | ✔        | ✗          | ✗        |
| Multi-error Support      | ✔        | ✗          | ~        |
| Fluid Interface          | ✔        | ✗          | ✗        |
| Go 1.13+ errors.Is/As    | ✔        | ✔          | ✔        |

## Best Practices

1. **Add Context Early**: Include relevant context information as close to the error source as possible
2. **Use Descriptive Messages**: Write clear, actionable error messages
3. **Preserve Original Errors**: Always wrap rather than replace original errors
4. **Structure Your Errors**: Use consistent patterns for error creation across your application
5. **Consider Your Audience**: Format errors appropriately for end users vs. developers

## License

This project is licensed under the terms of the MIT license.

[1]: https://github.com/reconquest/karma-go
[2]: https://godoc.org/github.com/pkg/errors
[3]: https://github.com/goph/emperror
