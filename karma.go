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
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
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

var (
	branchIndentation = "   "
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
	Context *Context        `json:"context,omitempty"`
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
			name+": "+ContextValueFormatter(value),
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
				stringReason(karma.Reason),
				"\n",
				"\n"+getBranchIndentation(),
				-1,
			)
	}
}

func getBranchIndentation() string {
	if len(branchIndentation) != BranchIndent {
		branchIndentation = strings.Repeat(" ", BranchIndent)
	}
	return branchIndentation
}

func stringReason(reason Reason) string {
	switch typed := reason.(type) {
	case []byte:
		return string(typed)

	case string:
		return typed

	case error:
		return typed.Error()

	default:
		return fmt.Sprint(typed)
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
func (karma Karma) Descend(callback func(Reason)) {
	// Do not descend into trivial cases, when message is reason, e.g. after
	// Reason() call.
	if karma.Message == "" {
		return
	}

	for _, reason := range karma.GetReasons() {
		switch reason := reason.(type) {
		case Karma:
			callback(reason)
			reason.Descend(callback)
		default:
			callback(reason)
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

	newReasons := parent.GetReasons()

	for _, reason := range reasons {
		if reason != nil {
			newReasons = append(newReasons, reason)
		}
	}

	return Karma{
		Message: parent.Message,
		Reason:  newReasons,
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

// Find typed object in given chain of reasons, returns true if reason with the
// same type found, if typed object is addressable, value will be stored in it.
func Find(err Reason, typed interface{}) bool {
	indirect := reflect.Indirect(reflect.ValueOf(typed))
	indirectType := indirect.Type()

	karma, ok := getKarma(err)
	if ok {
		return find(karma, typed, indirect, indirectType)
	}

	same := reflect.TypeOf(err) == indirectType
	if same {
		if indirect.CanAddr() {
			indirect.Set(reflect.ValueOf(err))
		}
	}

	return same
}

func find(
	karma *Karma,
	typed interface{},
	indirect reflect.Value,
	indirectType reflect.Type,
) bool {
	for _, nested := range karma.GetReasons() {
		subkarma, ok := getKarma(nested)
		if ok {
			if find(subkarma, typed, indirect, indirectType) {
				return true
			}
		} else {
			same := reflect.TypeOf(nested) == indirectType
			if same {
				if indirect.CanAddr() {
					indirect.Set(reflect.ValueOf(nested))
				}
			}

			return same
		}
	}

	return false
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

	return stringReason(chain) == stringReason(branch)
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
	message := bytes.NewBufferString(karma.Message)

	prolongate := false
	for _, reason := range reasons {
		if reasons, ok := reason.(Hierarchical); ok {
			if len(reasons.GetReasons()) > 0 {
				prolongate = true
				break
			}
		}
	}

	var (
		chainerLength = len([]rune(BranchChainer))
		splitter      = BranchSplitter
		chainer       = BranchChainer

		prolongator = "\n" + strings.TrimRightFunc(
			chainer, unicode.IsSpace,
		)
	)

	indentation := chainer
	if BranchIndent >= chainerLength {
		indentation += strings.Repeat(" ", BranchIndent-chainerLength)
	}

	for index, reason := range reasons {
		if index == len(reasons)-1 {
			splitter = BranchDelimiter
			if chainerLength < BranchIndent {
				chainerLength = BranchIndent
			}

			indentation = strings.Repeat(" ", chainerLength)
		}

		if message.Len() != 0 {
			message.WriteString("\n")
			message.WriteString(splitter)
		}

		message.WriteString(strings.Replace(
			stringReason(reason),
			"\n",
			"\n"+indentation,
			-1,
		))

		if prolongate && index < len(reasons)-1 {
			message.WriteString(prolongator)
		}
	}

	return message.String()
}
