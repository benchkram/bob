package envutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeEnv(t *testing.T) {
	type test struct {
		description    string
		first          []string
		second         []string
		expectedResult []string
	}

	tests := []test{
		{
			description:    "Merging two distinct env variables",
			first:          []string{"VAR_ONE=foo"},
			second:         []string{"VAR_TWO=bar"},
			expectedResult: []string{"VAR_ONE=foo", "VAR_TWO=bar"},
		},
		{
			description:    "Merging two variables with same key",
			first:          []string{"VAR_ONE=foo"},
			second:         []string{"VAR_ONE=bar"},
			expectedResult: []string{"VAR_ONE=bar"},
		},
		{
			description:    "With same variable in list",
			first:          []string{"VAR_ONE=foo", "VAR_TWO=xyz"},
			second:         []string{"VAR_ONE=bar", "HOME=/home/user"},
			expectedResult: []string{"VAR_TWO=xyz", "VAR_ONE=bar", "HOME=/home/user"},
		},
	}

	for _, tc := range tests {
		t.Log(tc.description)
		assert.Equal(t, tc.expectedResult, Merge(tc.first, tc.second))
	}
}
