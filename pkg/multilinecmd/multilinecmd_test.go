package multilinecmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultilineCmd(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "\n",
			expected: []string{},
		},
		{
			input:    " ",
			expected: []string{},
		},
		{
			input: "echo Hello",
			expected: []string{
				"echo Hello",
			},
		},
		{
			input: "  leading  and  trailing  spaces  ",
			expected: []string{
				"leading  and  trailing  spaces",
			},
		},
		{
			input: strings.Join([]string{
				"echo Hello",
				"some long\\\ncommand\\\nwith multiple line-breaks",
			}, "\n"),
			expected: []string{
				"echo Hello",
				"some long command with multiple line-breaks",
			},
		},
	}

	for _, test := range tests {
		result := Split(test.input)

		assert.Equal(t, len(test.expected), len(result))
		for i := range result {
			assert.Equal(t, test.expected[i], result[i], fmt.Sprintf("\nexpected: %s\ngot: %s\n", test.expected[i], result[i]))
		}
	}
}
