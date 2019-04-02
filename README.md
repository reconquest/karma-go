# Contexts

karma has contexts support, which allows to add arbitrary key-value fields
to the error to ease debug.

Simplest usage is to add key-values for existing error:

```go
func bar(arg string) error {
    return fmt.Errorf("arg is invalid: %s", arg)
}

func foo(arg string) error {
    err := bar(arg)
    if err != nil {
        return karma.Describe("method", "bar").Reason(err)
    }
}

fmt.Println(foo("hello"))

// Output:
//
// arg is invalid: hello
// └─ method: bar
```

# Comparison

| Feature                  | [karma][1] | [errors][2] | [emperror][3] |
| --                       | --         | --          | --            |
| Nested Errors            | ✔          | ✔           | ✔             |
| Key-Value Context        | ✔          |             | ~             |
| Descriptive Pretty Print | ✔          |             |               |
| Embedded Stack Trace     |            | ✔           | ✔             |
| JSON Friendly            | ✔          |             |               |
| Multi-error Support      | ✔          |             | ~             |
| Fluid Interface          | ✔          |             |               |

# License

This project is licensed under the terms of the MIT license.

[1]: https://github.com/reconquest/karma-go
[2]: https://godoc.org/github.com/pkg/errors
[3]: https://github.com/goph/emperror
