// Package karma provides a simple way to return and display hierarchical
// messages or errors.
//
// Transforms:
//
//         can't pull remote 'origin': can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*': exit status 128
//
// Into:
//
//         can't pull remote 'origin'
//         └─ can't run git fetch 'origin' 'refs/tokens/*:refs/tokens/*'
//            └─ exit status 128
package karma // import "github.com/reconquest/karma-go"

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const (
	// BranchDelimiterASCII represents a simple ASCII delimiter for hierarchy
	// branches.
	//
	// Use: karma.BranchDelimiter = karma.BranchDelimiterASCII
	BranchDelimiterASCII = `\_ `

	// BranchDelimiterBox represents UTF8 delimiter for hierarchy branches.
	//
	// Use: karma.BranchDelimiter = karma.BranchDelimiterBox
	BranchDelimiterBox = `└─ `

	// BranchChainerASCII represents a simple ASCII chainer for hierarchy
	// branches.
	//
	// Use: karma.BranchChainer = karma.BranchChainerASCII
	BranchChainerASCII = `| `

	// BranchChainerBox represents UTF8 chainer for hierarchy branches.
	//
	// Use: karma.BranchChainer = karma.BranchChainerBox
	BranchChainerBox = `│ `

	// BranchSplitterASCII represents a simple ASCII splitter for hierarchy
	// branches.
	//
	// Use: karma.BranchSplitter = karma.BranchSplitterASCII
	BranchSplitterASCII = `+ `

	// BranchSplitterBox represents UTF8 splitter for hierarchy branches.
	//
	// Use: karma.BranchSplitter = karma.BranchSplitterBox
	BranchSplitterBox = `├─ `
)

var (
	// BranchDelimiter set delimiter each nested message text will be started
	// from.
	BranchDelimiter = BranchDelimiterBox

	// BranchChainer set chainer each nested message tree text will be started
	// from.
	BranchChainer = BranchChainerBox

	// BranchSplitter set splitter each nested messages splitted by.
	BranchSplitter = BranchSplitterBox

	// BranchIndent set number of spaces each nested message will be indented by.
	BranchIndent = 3
)

// Karma represents hierarchy message, linked with nested message.
type Karma struct {
	// Reason of message, which can be Karma as well.
	Reason Reason

	// Message is formatted message, which will be returned when String()
	// will be invoked.
	Message string

	// Context is a key-pair linked list, which represents runtime context
	// of the situtation.
	Context *Context
}

// Hierarchical represents interface, which methods will be used instead
// of calling String() and Karma() methods.
type Hierarchical interface {
	// String returns hierarchical string representation.
	String() string

	// GetReasons returns slice of nested reasons.
	GetReasons() []Reason

	// GetMessage returns top-level message.
	GetMessage() string
}

// Reason is either `error` or string.
type Reason interface{}

type jsonRepresentation struct {
	Reason  json.RawMessage `json:"reason,omitempty"`
	Message string          `json:"message,omitempty"`
	Context *Context        `json:"context"`
}

// Format creates new hierarchical message.
//
// With reason == nil call will be equal to `fmt.Errorf()`.
func Format(
	reason Reason,
	message string,
	args ...interface{},
) Karma {
	return Karma{
		Message: fmt.Sprintf(message, args...),
		Reason:  reason,
	}
}

// ContextValueFormatter returns string representation of context value when
// Format() is called on Karma struct.
var ContextValueFormatter = func(value interface{}) string {
	if value, ok := value.(string); ok {
		if value == "" {
			return "<empty>"
		}
	}

	return fmt.Sprint(value)
}

// Karma returns hierarchical string representation. If no nested
// message was specified, then only current message will be returned.
func (karma Karma) String() string {
	karma.Context.Walk(func(name string, value interface{}) {
		karma = Push(karma, Push(
			fmt.Sprintf("%s: %s", name, ContextValueFormatter(value)),
		))
	})

	switch value := karma.Reason.(type) {
	case nil:
		return karma.Message

	case []Reason:
		return formatReasons(karma, value)

	default:
		return karma.Message + "\n" +
			BranchDelimiter +
			strings.Replace(
				fmt.Sprintf("%s", karma.Reason),
				"\n",
				"\n"+strings.Repeat(" ", BranchIndent),
				-1,
			)
	}
}

// Error implements error interface, Karma can be returned as error.
func (karma Karma) Error() string {
	return karma.String()
}

// GetReasons returns nested messages, embedded into message.
func (karma Karma) GetReasons() []Reason {
	if karma.Reason == nil {
		return nil
	}

	if reasons, ok := karma.Reason.([]Reason); ok {
		return reasons
	} else {
		return []Reason{karma.Reason}
	}
}

// GetMessage returns message message
func (karma Karma) GetMessage() string {
	if karma.Message == "" {
		return fmt.Sprintf("%s", karma.Reason)
	} else {
		return karma.Message
	}
}

// GetContext returns context
func (karma Karma) GetContext() *Context {
	return karma.Context
}

// Descend calls specified callback for every nested hierarchical message.
func (karma Karma) Descend(callback func(Karma)) {
	for _, reason := range karma.GetReasons() {
		if reason, ok := reason.(Karma); ok {
			callback(reason)

			reason.Descend(callback)
		}
	}
}

func (karma Karma) MarshalJSON() ([]byte, error) {
	result := jsonRepresentation{
		Message: karma.Message,
		Context: karma.Context,
	}

	var err error

	switch reason := karma.Reason.(type) {
	case json.Marshaler:
		result.Reason, err = json.Marshal(reason)
	case error:
		result.Reason, err = json.Marshal(reason.Error())
	default:
		result.Reason, err = json.Marshal(reason)
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (karma *Karma) UnmarshalJSON(data []byte) error {
	var container jsonRepresentation

	err := json.Unmarshal(data, &container)
	if err != nil {
		return err
	}

	var reason Karma

	if len(container.Reason) > 0 {
		err = json.Unmarshal(container.Reason, &reason)
		if err != nil {
			err = json.Unmarshal(container.Reason, &karma.Reason)
			if err != nil {
				return err
			}
		} else {
			karma.Reason = reason
		}
	}

	karma.Message = container.Message
	karma.Context = container.Context

	return nil
}

// Push creates new hierarchy message with multiple branches separated by
// separator, delimited by delimiter and prolongated by prolongator.
func Push(reason Reason, reasons ...Reason) Karma {
	parent, ok := reason.(Karma)
	if !ok {
		parent = Karma{
			Message: fmt.Sprint(reason),
		}
	}

	return Karma{
		Message: parent.Message,
		Reason:  append(parent.GetReasons(), reasons...),
	}
}

// Describe creates new context list, which can be used to produce context-rich
// hierarchical message.
func Describe(key string, value interface{}) *Context {
	return &Context{
		KeyValue: KeyValue{
			Key:   key,
			Value: value,
		},
	}
}

// Contains returns true when branch is found in reasons of given chain. Or
// chain has the same value as branch error.
// Useful when you work with result of multi-level error and just wanted to
// check that error contains os.ErrNoExist.
func Contains(chain Reason, branch Reason) bool {
	karma, ok := getKarma(chain)
	if ok {
		return contains(karma, branch)
	}

	return fmt.Sprint(chain) == fmt.Sprint(branch)
}

func contains(karma *Karma, reason Reason) bool {
	reasonString := fmt.Sprint(reason)
	for _, nested := range karma.GetReasons() {
		subkarma, ok := getKarma(nested)
		if ok {
			if contains(subkarma, reason) {
				return true
			}
		} else {
			if fmt.Sprint(nested) == reasonString {
				return true
			}
		}
	}

	return false
}

func getKarma(reason Reason) (*Karma, bool) {
	karma, ok := reason.(Karma)
	if ok {
		return &karma, true
	}

	pointer, ok := reason.(*Karma)
	if ok {
		return pointer, true
	}

	return nil, false
}

func formatReasons(karma Karma, reasons []Reason) string {
	message := karma.Message

	prolongate := false
	for _, reason := range reasons {
		if reasons, ok := reason.(Hierarchical); ok {
			if len(reasons.GetReasons()) > 0 {
				prolongate = true
				break
			}
		}
	}

	for index, reason := range reasons {
		var (
			splitter      = BranchSplitter
			chainer       = BranchChainer
			chainerLength = len([]rune(BranchChainer))
		)

		if index == len(reasons)-1 {
			splitter = BranchDelimiter
			chainer = strings.Repeat(" ", chainerLength)
		}

		indentation := chainer
		if BranchIndent >= chainerLength {
			indentation += strings.Repeat(" ", BranchIndent-chainerLength)
		}

		prolongator := ""
		if prolongate && index < len(reasons)-1 {
			prolongator = "\n" + strings.TrimRightFunc(
				chainer, unicode.IsSpace,
			)
		}

		if message != "" {
			message = message + "\n" + splitter
		}

		message += strings.Replace(
			fmt.Sprint(reason),
			"\n",
			"\n"+indentation,
			-1,
		)
		message += prolongator
	}

	return message
}
