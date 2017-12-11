package karma

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat_CanFormatEmptyError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Format(nil, ""), "")
}

func TestFormat_CanFormatSimpleStringError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Format(nil, "simple error"), "simple error")
}

func TestFormat_CanFormatSimpleStringErrorWithArgs(t *testing.T) {
	test := assert.New(t)

	test.EqualError(Format(nil, "integer: %d", 9), "integer: 9")
}

func TestFormat_CanFormatErrorWithSimpleReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format(errors.New("reason"), "everything has a reason"),
		output(
			"everything has a reason",
			"└─ reason",
		),
	)
}

func TestFormat_CanFormatErrorWithSimpleReasonAndArgs(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format(errors.New("reason"), "reasons: %d", 1),
		output(
			"reasons: 1",
			"└─ reason",
		),
	)
}

func TestFormat_CanFormatHierarchicalReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format(Format(errors.New("reason"), "cause"), "karma"),
		output(
			"karma",
			"└─ cause",
			"   └─ reason",
		),
	)
}

func TestFormat_CanFormatHierarchicalReasonWithSimpleReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format(Format("reason", "cause"), "karma"),
		output(
			"karma",
			"└─ cause",
			"   └─ reason",
		),
	)
}

func TestFormat_CanFormatAnyReason(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format([]byte("self"), "no"),
		output(
			"no",
			"└─ self",
		),
	)
}

func TestCanSetBranchDelimiter(t *testing.T) {
	test := assert.New(t)

	delimiter := BranchDelimiter
	defer func() {
		BranchDelimiter = delimiter
	}()

	BranchDelimiter = "* "

	test.EqualError(
		Format(Format("first", "second"), "third"),
		output(
			"third",
			"* second",
			"   * first",
		),
	)
}

func TestCanSetBranchIndent(t *testing.T) {
	test := assert.New(t)

	indent := BranchIndent
	defer func() {
		BranchIndent = indent
	}()

	BranchIndent = 0

	test.EqualError(
		Format(Format("first", "second"), "third"),
		output(
			"third",
			"└─ second",
			"└─ first",
		),
	)
}

func TestCanMarshalToJSON(t *testing.T) {
	test := assert.New(t)

	item := Describe("host", "example.com").Format(
		Describe("os", "linux").Reason(
			"system error",
		),
		"unable to resolve",
	)

	marshalled, err := json.MarshalIndent(item, ``, `  `)
	test.NoError(err)
	test.JSONEq(string(marshalled), `{
		  "reason": {
			"reason": "system error",
			"context": [
			  {
				"key": "os",
				"value": "linux"
			  }
			]
		  },
		  "message": "unable to resolve",
		  "context": [
			{
			  "key": "host",
			  "value": "example.com"
			}
		  ]
		}`,
	)
}

func TestCanMarshalErrorsToJSON(t *testing.T) {
	test := assert.New(t)

	item := Describe("host", "example.com").Format(
		errors.New("access denied"),
		"unable to connect",
	)

	marshalled, err := json.MarshalIndent(item, ``, `  `)
	test.NoError(err)
	test.JSONEq(string(marshalled), `{
		  "reason": "access denied",
		  "message": "unable to connect",
		  "context": [
			{
			  "key": "host",
			  "value": "example.com"
			}
		  ]
		}`,
	)
}

func TestCanUnmarshalFromJSON(t *testing.T) {
	return
	test := assert.New(t)

	input := `{
		  "reason": "access denied",
		  "message": "unable to connect",
		  "context": [
			{
			  "key": "host",
			  "value": "example.com"
			}
		  ]
		}`

	var actual Karma

	err := json.Unmarshal([]byte(input), &actual)
	test.NoError(err)

	test.EqualError(
		actual,
		output(
			"unable to connect",
			"├─ access denied",
			"└─ host: example.com",
		),
	)
}

func TestCanUnmarshalNestedReasonFromJSON(t *testing.T) {
	test := assert.New(t)

	input := `{
		  "reason": {
			  "message": "tcp: out of memory",
			  "context": [
				{
				  "key": "free",
				  "value": "512Kb"
				}
			  ]
		  },
		  "message": "unable to connect",
		  "context": [
			{
			  "key": "host",
			  "value": "example.com"
			}
		  ]
		}`

	var actual Karma

	err := json.Unmarshal([]byte(input), &actual)
	test.NoError(err)

	test.EqualError(
		actual,
		output(
			"unable to connect",
			"├─ tcp: out of memory",
			"│  └─ free: 512Kb",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanAddMultipleKeyValues(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("host", "example.com").Describe("operation", "resolv").Format(
			"system error",
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"├─ host: example.com",
			"└─ operation: resolv",
		),
	)
}

func TestContext_CanAddWithoutHierarchy(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("host", "example.com").Reason(
			Describe("operation", "resolv").Reason(
				"system error",
			),
		),
		output(
			"system error",
			"├─ operation: resolv",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanAddToRootError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("host", "example.com").Format(
			"system error",
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanAddToReasonError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("host", "example.com").Format(
			Describe("os", "linux").Reason(
				"system error",
			),
			"unable to resolve",
		),
		output(
			"unable to resolve",
			"├─ system error",
			"│  └─ os: linux",
			"│",
			"└─ host: example.com",
		),
	)
}

func TestContext_CanUseNonStringValue(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("code", 88).Format(
			nil,
			"unable to run external command",
		),
		output(
			"unable to run external command",
			"└─ code: 88",
		),
	)
}

func TestContext_DontChangeSelf(t *testing.T) {
	test := assert.New(t)

	void := Describe("void", 0).Describe("emptiness", 0)

	test.EqualError(
		void.Describe("space", 1).Format(nil, "the story"),
		output(
			"the story",
			"├─ void: 0",
			"├─ emptiness: 0",
			"└─ space: 1",
		),
	)

	test.EqualError(
		void.Describe("time", 1).Format(nil, "the story"),
		output(
			"the story",
			"├─ void: 0",
			"├─ emptiness: 0",
			"└─ time: 1",
		),
	)
}

func TestContext_FieldsNotSorted(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Describe("start_time", 0).
			Describe("end_time", 1).
			Describe("precision", 2).
			Describe("offset", 3).
			Format(
				nil,
				"fields are not sorted",
			),
		output(
			"fields are not sorted",
			"├─ start_time: 0",
			"├─ end_time: 1",
			"├─ precision: 2",
			"└─ offset: 3",
		),
	)
}

func TestContext_CanOperateOnNilContext(t *testing.T) {
	test := assert.New(t)

	var void *Context

	void = void.Describe("space", 1)
	void = void.Describe("time", 2)

	test.EqualError(
		void.Format(nil, "the story"),
		output(
			"the story",
			"├─ space: 1",
			"└─ time: 2",
		),
	)
}

type customError struct {
	Text   string
	Reason error
}

func (err customError) Error() string {
	return Format(err.Reason, err.GetMessage()).Error()
}

func (err customError) GetNested() []Reason {
	return []Reason{err.Reason}
}

func (err customError) GetMessage() string {
	return strings.ToUpper(err.Text)
}

func TestCustomHierarchicalError(t *testing.T) {
	test := assert.New(t)

	test.EqualError(
		Format(
			customError{"upper", errors.New("hierarchical")},
			"example of custom error",
		),
		output(
			"example of custom error",
			"└─ UPPER",
			"   └─ hierarchical",
		),
	)
}

func ExampleContext_MultipleKeyValues() {
	foo := func(arg string) error {
		return fmt.Errorf("unable to foo on %s", arg)
	}

	bar := func() error {
		err := foo("zen")
		if err != nil {
			return Describe("method", "foo").Describe("arg", "zen").Reason(err)
		}

		return nil
	}

	err := bar()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to foo on zen
	// ├─ method: foo
	// └─ arg: zen
}

func ExampleContext_NestedErrors() {
	foo := func(arg string) error {
		return fmt.Errorf("unable to foo on %s", arg)
	}

	bar := func() error {
		err := foo("zen")
		if err != nil {
			return Describe("arg", "zen").Reason(err)
		}

		return nil
	}

	baz := func() error {
		err := bar()
		if err != nil {
			return Describe("operation", "foo").Format(
				err,
				"unable to perform critical operation",
			)
		}

		return nil
	}

	err := baz()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to perform critical operation
	// ├─ unable to foo on zen
	// │  └─ arg: zen
	// │
	// └─ operation: foo
}

func ExampleContext_AddNestedDescribe() {
	foo := func() error {
		return fmt.Errorf("unable to foo")
	}

	bar := func() error {
		err := foo()
		if err != nil {
			return Describe("level", "bar").Reason(err)
		}

		return nil
	}

	baz := func() error {
		err := bar()
		if err != nil {
			return Describe("level", "baz").Reason(err)
		}

		return nil
	}

	err := baz()
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	//
	// unable to foo
	// ├─ level: bar
	// └─ level: baz
}

func ExampleContext_UseCustomLoggingFormat() {
	// solve function represents deepest function in the call stack
	solve := func(koan string) error {
		return fmt.Errorf("no solution available for %q", koan)
	}

	// think represents function, which calls solve function
	think := func() error {
		err := solve("what was your face before your parents were born?")
		if err != nil {
			return Describe("though", "koan").Reason(err)
		}

		return nil
	}

	// realize represents top-level function, which calls think function
	realize := func() error {
		context := Describe("doing", "realization")

		err := think()
		if err != nil {
			return context.Describe("action", "thinking").Format(
				err,
				"unable to attain realization",
			)
		}

		return nil
	}

	// log represents custom logging function, which writes structured logs,
	// like logrus in format [LEVEL] message: key1=value1 key2=value2
	log := func(level string, message string, kv ...interface{}) {
		fmt.Printf("[%s] %s:", level, message)

		for i := 0; i < len(kv); i += 2 {
			fmt.Printf(" %s=%q", kv[i], kv[i+1])
		}

		fmt.Println()
	}

	err := realize()
	if err != nil {
		if karma, ok := err.(Karma); ok {
			// following call will write all nested errors
			karma.Descend(func(karma Karma) {
				log(
					"ERROR",
					karma.GetMessage(),
					karma.GetContext().GetKeyValuePairs()...,
				)
			})

			// this call will write only root-level error
			log(
				"FATAL",
				karma.GetMessage(),
				karma.GetContext().GetKeyValuePairs()...,
			)
		}
	}

	// Output:
	//
	// [ERROR] no solution available for "what was your face before your parents were born?": though="koan"
	// [FATAL] unable to attain realization: doing="realization" action="thinking"
}

func output(lines ...string) string {
	return strings.Join(lines, "\n")
}
