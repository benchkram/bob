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
		{input: "../strange/target", want: false},
		{input: "/", want: false},

		// valid
		{input: "my/../strange/target", want: true},
		{input: "target", want: true},
		{input: "./target", want: true},
		{input: "./target/file..go", want: true},
	}

	for _, tc := range tests {
		got := isValidFilesystemTarget(tc.input)

		if !assert.Equal(t, tc.want, got) {
			t.Fatalf("input %v,  expected: %t but is %t ", tc.input, tc.want, got)
		}
	}
}
