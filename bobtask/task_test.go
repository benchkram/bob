package bobtask

import (
	"testing"

	"github.com/benchkram/bob/pkg/reflectutil"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

var withLowercase = `
type: binary
path: ./build/server
dependson:
      - build
      - database
`

var withCamelCase = `
type: binary
path: ./build/server
dependsOn:
      - build
      - database
`

var withBoth = `
type: binary
path: ./build/server
dependsOn:
      - build
      - database
dependson:
      - build
      - database
`

func TestTask_UnmarshalYAMLDependsOn(t *testing.T) {
	type test struct {
		input string
		msg   string
	}

	tests := []test{
		{input: withLowercase, msg: "Should equal values from `dependson`"},
		{input: withCamelCase, msg: "Should equal values from `dependsOn`"},
	}

	for _, tc := range tests {
		var task Task

		err := yaml.Unmarshal([]byte(tc.input), &task)

		assert.Nil(t, err, "No error should occur on Unmarshal")
		assert.Equal(t, []string{"build", "database"}, task.DependsOn, tc.msg)
	}
}

func TestTask_UnmarshalYAMLWithBothDependsOn(t *testing.T) {
	t.Log("When both values exists for a task should fail with error")

	var task Task
	err := yaml.Unmarshal([]byte(withBoth), &task)
	assert.EqualError(t, err, "both `dependson` and `dependsOn` nodes detected near line 2")
}

// This test was added to make sure that IsValidDecoration from Task
// is validating all exported fields.
//
// When this test fails you should add the new/remove field to IsValidDecoration
// and then update the test with correct expected number
func TestTask_NumberOfPublicFields(t *testing.T) {
	var task Task
	expected := 7
	got := reflectutil.CountExportedFields(task)

	if got != expected {
		t.Errorf("Expected %d, got %d", expected, got)
	}
}
