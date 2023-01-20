package fib

import (
	"context"
	"os"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/test/setup"
	"github.com/stretchr/testify/assert"
)

var result error

func BenchmarkBob(b *testing.B) {
	dir, storageDir, cleanup, err := setup.TestDirs("build-benchmark")
	assert.Nil(b, err)
	defer cleanup()

	err = os.Chdir(dir)
	assert.Nil(b, err)

	bobInstance, err := bob.BobWithBaseStoreDir(storageDir, bob.WithDir(dir))
	assert.Nil(b, err)

	err = bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})
	assert.Nil(b, err)

	ctx := context.Background()
	err = bobInstance.Build(ctx, bob.BuildAllTargetName)
	assert.Nil(b, err)

	var r error
	for n := 0; n < b.N; n++ {
		// always record the result of Fib to prevent
		// the compiler eliminating the function call.
		r = bobInstance.Build(ctx, bob.BuildAllTargetName)

	}
	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	result = r

	b.ReportAllocs()
}
