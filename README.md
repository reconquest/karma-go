# Contexts

karma has contexts support, which allows to add arbitrary key-value fields
to the error to ease debug.

Simplest usage is to add key-values for existing error:

```
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

# License

This project is licensed under the terms of the MIT license.
