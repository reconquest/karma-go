package karma_test

import (
	"fmt"
	"io"

	"github.com/reconquest/karma-go"
)

func foo(a int) error {
	return karma.Format(io.EOF, "eof or something")
}

func bar(quz int) error {
	wox := quz * 2

	err := foo(wox)
	if err != nil {
		return karma.
			Describe("wox", wox).
			Format(err, "foo")
	}

	return nil
}

func twix() error {
	err := bar(42)
	if err != nil {
		return karma.Describe("barval", 42).Format(err, "bar")
	}

	return nil
}

func run() error {
	err := twix()
	if err != nil {
		return karma.Format(err, "twix")
	}

	return nil
}

func ExampleFlatten() {
	{
		err := io.EOF
		fmt.Println("regular:")
		fmt.Println(err)
		fmt.Println("flatten:")
		fmt.Println(karma.Flatten(err))
	}

	{
		err := run()
		fmt.Println("regular (hierarchical):")
		fmt.Println(err)
		fmt.Println("flatten:")
		fmt.Println(karma.Flatten(err))
	}

	// Output:
	//
	// regular:
	// EOF
	// flatten:
	// EOF
	// regular (hierarchical):
	// twix
	// └─ bar
	//    ├─ foo
	//    │  ├─ eof or something
	//    │  │  └─ EOF
	//    │  │
	//    │  └─ wox: 84
	//    │
	//    └─ barval: 42
	// flatten:
	// twix: bar: foo: eof or something: EOF | barval=42 wox=84
}
