package bob

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {

	type test struct {
		input    []string
		expected []string
	}

	tests := []test{

		{
			// 1 - 2 - 3
			input:    []string{"1", "2", "3"},
			expected: []string{"1", "2", "3"},
		},

		{
			// 1 - 2 - 3
			//   \
			//     3 - 4
			input:    []string{"1", "2", "3", "3", "4"},
			expected: []string{"1", "2", "3", "4"},
		},

		{
			// 1 - 2 - 5 - 4
			// | \
			// |   3 - 4
			//  \
			//    5 - 4
			input:    []string{"1", "2", "5", "4", "3", "4", "5", "4"},
			expected: []string{"1", "2", "3", "5", "4"},
		},
	}

	for _, test := range tests {
		result := normalize(test.input)
		assert.True(t, reflect.DeepEqual(test.expected, result))
	}
}
