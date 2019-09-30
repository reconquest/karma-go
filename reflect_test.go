package karma

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReflect(t *testing.T) {
	test := assert.New(t)

	delimiter := BranchDelimiter
	defer func() {
		BranchDelimiter = delimiter
	}()

	BranchDelimiter = "* "

	type Quo struct {
		Q bool
		w string
	}
	foo := struct {
		Bar struct {
			SliceStrings []string
			SimpleStruct struct {
				Int int
			}
		}
		SliceStructs []Quo
		Map          map[string][]int
	}{}

	foo.Bar.SliceStrings = []string{"a", "b"}
	foo.Bar.SimpleStruct.Int = 11
	foo.SliceStructs = append(foo.SliceStructs, Quo{
		Q: true,
		w: "w1",
	})
	foo.SliceStructs = append(foo.SliceStructs, Quo{
		Q: false,
		w: "w",
	})
	foo.Map = map[string][]int{
		"a": []int{1, 2},
		"b": []int{3, 4},
	}

	values := DescribeDeep("foo", foo).GetKeyValuePairs()
	chunks := []string{}
	for i := 0; i < len(values); i += 2 {
		if i == 0 {
			continue
		}
		chunks = append(chunks, fmt.Sprintf("%s=%v", values[i], values[i+1]))
	}

	test.EqualValues(
		[]string{
			"foo.Bar.SliceStrings[0]=a",
			"foo.Bar.SliceStrings[1]=b",
			"foo.Bar.SimpleStruct.Int=11",
			"foo.SliceStructs[0].Q=true",
			"foo.SliceStructs[1].Q=false",
			"foo.Map=map[a:[1 2] b:[3 4]]",
		},
		chunks,
	)
}
