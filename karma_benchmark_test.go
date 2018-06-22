package karma

import (
	"errors"
	"fmt"
	"testing"
)

func BenchmarkKarmaFormat(b *testing.B) {
	nested := errors.New("nested")
	for i := 0; i < b.N; i++ {
		err := Format(nested, "parent")
		_ = err
	}
}

func BenchmarkFmtErrorf(b *testing.B) {
	nested := errors.New("nested")
	for i := 0; i < b.N; i++ {
		err := fmt.Errorf("parent: %s", nested)
		_ = err
	}
}

func BenchmarkKarmaFormat_String(b *testing.B) {
	nested := errors.New("nested")
	for i := 0; i < b.N; i++ {
		err := Format(nested, "parent")
		_ = err.Error()
	}
}

func BenchmarkFmtErrorf_String(b *testing.B) {
	nested := errors.New("nested")
	for i := 0; i < b.N; i++ {
		err := fmt.Errorf("parent: %s", nested)
		_ = err.Error()
	}
}
