package bob

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "net/http/pprof"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/stretchr/testify/assert"
)

var result *bobfile.Bobfile

func BenchmarkAggregateOnPlayground(b *testing.B) {

	b.StopTimer() // Stop during initialization

	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-benchmark-playground-*")
	assert.Nil(b, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(b, err)

	testBob, err := Bob(WithDir(dir))
	assert.Nil(b, err)

	err = CreatePlayground(dir, "")
	assert.Nil(b, err)

	var bobfile *bobfile.Bobfile // to block compiler optimization

	// actual benchmark
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		bobfile, err = testBob.Aggregate()

	}
	// avoid counting the cleanup
	b.StopTimer()
	assert.Nil(b, err)
	// "store" bobfile
	result = bobfile // to block compiler optimization

}

func BenchmarkAggregate(b *testing.B)     { benchmarkAggregate(b, 0) }
func BenchmarkAggregate10(b *testing.B)   { benchmarkAggregate(b, 10) }
func BenchmarkAggregate20(b *testing.B)   { benchmarkAggregate(b, 20) }
func BenchmarkAggregate50(b *testing.B)   { benchmarkAggregate(b, 50) }
func BenchmarkAggregate100(b *testing.B)  { benchmarkAggregate(b, 100) }
func BenchmarkAggregate1000(b *testing.B) { benchmarkAggregate(b, 1000) }
func BenchmarkAggregate2000(b *testing.B) { benchmarkAggregate(b, 2000) }

const defaultMultiplier = 200

func benchmarkAggregate(b *testing.B, ignoredMultiplier int) {

	b.StopTimer() // Stop during initialization

	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-benchmark-*")
	assert.Nil(b, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(b, err)

	testBob, err := Bob(WithDir(dir))
	assert.Nil(b, err)

	err = CreatePlayground(dir, "")
	assert.Nil(b, err)

	// Create a file structure  which Aggregate will completly travers
	err = createFileSturcture(dir, defaultMultiplier)
	assert.Nil(b, err)

	err = createIgnoreFileSturcture(dir, ignoredMultiplier)
	assert.Nil(b, err)

	var bobfile *bobfile.Bobfile // to block compiler optimization

	// actual benchmark
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		bobfile, err = testBob.Aggregate()

	}
	// avoid counting the cleanup
	b.StopTimer()
	assert.Nil(b, err)
	// "store" bobfile
	result = bobfile // to block compiler optimization

}

// createFileSturcture creates a deep file structure.
// `multiplier`` is the number of directorys created containing the structure.
func createFileSturcture(dir string, multiplier int) error {
	for i := 0; i < multiplier; i++ {
		// create parent
		err := os.MkdirAll(fmt.Sprintf("%s/%d/first/second/third/fourth", dir, i), os.ModePerm)
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/%d/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/%d/first/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/%d/first/second/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/%d/first/second/third/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/%d/first/second/third/fourth/a", dir, i))
		if err != nil {
			return err
		}
	}

	return nil
}

// createIgnoreFileSturcture creates a deep file structure witch must be ignored by Aggregate().
// `multiplier`` is the number of directorys created containing the structure.
func createIgnoreFileSturcture(dir string, multiplier int) error {
	for i := 0; i < multiplier; i++ {
		// create parent
		err := os.MkdirAll(fmt.Sprintf("%s/node_modules/%d/first/second/third/fourth", dir, i), os.ModePerm)
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/node_modules/%d/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/node_modules/%d/first/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/node_modules/%d/first/second/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/node_modules/%d/first/second/third/a", dir, i))
		if err != nil {
			return err
		}

		_, err = os.Create(fmt.Sprintf("%s/node_modules/%d/first/second/third/fourth/a", dir, i))
		if err != nil {
			return err
		}
	}

	return nil
}

func TestEmptyProjectName(t *testing.T) {
	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(t, err)

	testBob, err := Bob(WithDir(dir))
	assert.Nil(t, err)

	err = CreatePlayground(dir, "")
	assert.Nil(t, err)

	bobfile, err := testBob.Aggregate()
	assert.Nil(t, err)

	assert.Equal(t, dir, bobfile.Project)
}

func TestProjectName(t *testing.T) {
	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(t, err)

	testBob, err := Bob(WithDir(dir))
	assert.Nil(t, err)

	projectName := "example.com/test-user/test-project"

	err = CreatePlayground(dir, projectName)
	assert.Nil(t, err)

	bobfile, err := testBob.Aggregate()
	assert.Nil(t, err)

	assert.Equal(t, projectName, bobfile.Project)
}

func TestInvalidProjectName(t *testing.T) {
	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(t, err)

	testBob, err := Bob(WithDir(dir))
	assert.Nil(t, err)

	projectName := "{}"

	err = CreatePlayground(dir, projectName)
	assert.Nil(t, err)

	_, err = testBob.Aggregate()
	assert.ErrorIs(t, err, bobfile.ErrInvalidProjectName)
}
