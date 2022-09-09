package bobtask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidSystemTarget(t *testing.T) {
	type test struct {
		input string
		want  bool
	}

	tests := []test{
		// invalid
		{input: ".", want: false},
		{input: "..", want: false},
		{input: "...", want: false},
		{input: "./..", want: false},
		{input: "../target", want: false},
		{input: "my/../strange/target", want: false},
		// valid
		{input: "target", want: true},
		{input: "./target", want: true},
	}

	for _, tc := range tests {
		got := isValidFilesystemTarget(tc.input)

		if !assert.Equal(t, tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
