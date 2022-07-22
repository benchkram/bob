package bobrun

import (
	"testing"

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

func TestRun_UnmarshalYAMLDependsOn(t *testing.T) {
	type test struct {
		input string
		msg   string
	}

	tests := []test{
		{input: withLowercase, msg: "Should equal values from `dependson`"},
		{input: withCamelCase, msg: "Should equal values from `dependsOn`"},
	}

	for _, tc := range tests {
		var run Run

		err := yaml.Unmarshal([]byte(tc.input), &run)

		assert.Nil(t, err, "No error should occur on Unmarshal")
		assert.Equal(t, []string{"build", "database"}, run.DependsOn, tc.msg)
	}
}

func TestRun_UnmarshalYAMLWithBothDependsOn(t *testing.T) {
	t.Log("When both values exists for a run should fail with error")

	var run Run
	err := yaml.Unmarshal([]byte(withBoth), &run)
	assert.EqualError(t, err, "both `dependson` and `dependsOn` nodes detected near line 2")
}
