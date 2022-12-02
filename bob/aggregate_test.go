package bob

import (
	"fmt"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"testing"

	"github.com/benchkram/errz"
	"github.com/stretchr/testify/assert"

	"github.com/benchkram/bob/bob/bobfile"
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

	err = CreatePlayground(PlaygroundOptions{Dir: dir})
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

	err = CreatePlayground(PlaygroundOptions{Dir: dir})
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
// `multiplier“ is the number of directorys created containing the structure.
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
// `multiplier“ is the number of directorys created containing the structure.
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

	testBob, err := Bob(WithDir(dir), WithAllowRedundantTargets())
	assert.Nil(t, err)

	err = CreatePlayground(PlaygroundOptions{Dir: dir})
	assert.Nil(t, err)

	bobfile, err := testBob.Aggregate()
	errz.Log(err)
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

	testBob, err := Bob(WithDir(dir), WithAllowRedundantTargets())
	assert.Nil(t, err)

	projectName := "example.com/test-user/test-project"

	err = CreatePlayground(PlaygroundOptions{
		Dir:         dir,
		ProjectName: projectName,
	})
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

	testBob, err := Bob(WithDir(dir), WithAllowRedundantTargets())
	assert.Nil(t, err)

	projectName := "@"

	err = CreatePlayground(PlaygroundOptions{
		Dir:         dir,
		ProjectName: projectName,
	})
	assert.Nil(t, err)

	_, err = testBob.Aggregate()
	assert.ErrorIs(t, err, bobfile.ErrInvalidProjectName)
}

// FIXME: As we don't refer to a child task by projectname but by path
// it seems to be save to allow duplicate projectnames.
//
// func TestDuplicateProjectNameSimple(t *testing.T) {
// 	// Create playground
// 	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
// 	assert.Nil(t, err)

// 	defer os.RemoveAll(dir)

// 	err = os.Chdir(dir)
// 	assert.Nil(t, err)

// 	testBob, err := Bob(WithDir(dir))
// 	assert.Nil(t, err)

// 	projectName := "duplicated-name"
// 	projectNameSecondLevel := "duplicated-name"
// 	projectNameThirdLevel := "third-level"

// 	err = CreatePlayground(
// 		PlaygroundOptions{
// 			Dir:                    dir,
// 			ProjectName:            projectName,
// 			ProjectNameSecondLevel: projectNameSecondLevel,
// 			ProjectNameThirdLevel:  projectNameThirdLevel,
// 		},
// 	)
// 	assert.Nil(t, err)

// 	_, err = testBob.Aggregate()
// 	assert.ErrorIs(t, err, ErrDuplicateProjectName)
// }

// FIXME: As we don't refer to a child task by projectname but by path
// it seems to be save to allow duplicate projectnames.
//
// func TestDuplicateProjectNameComplex(t *testing.T) {
// 	// Create playground
// 	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
// 	assert.Nil(t, err)

// 	defer os.RemoveAll(dir)

// 	err = os.Chdir(dir)
// 	assert.Nil(t, err)

// 	testBob, err := Bob(WithDir(dir))
// 	assert.Nil(t, err)

// 	projectName := "bob.build/benchkram/duplicated-name"
// 	projectNameSecondLevel := "bob.build/benchkram/duplicated-name"
// 	projectNameThirdLevel := "bob.build/benchkram/third-level"

// 	err = CreatePlayground(
// 		PlaygroundOptions{
// 			Dir:                    dir,
// 			ProjectName:            projectName,
// 			ProjectNameSecondLevel: projectNameSecondLevel,
// 			ProjectNameThirdLevel:  projectNameThirdLevel,
// 		},
// 	)
// 	assert.Nil(t, err)

// 	_, err = testBob.Aggregate()
// 	assert.ErrorIs(t, err, ErrDuplicateProjectName)
// }

func TestMultiLevelBobfileSameProjectName(t *testing.T) {
	// Create playground
	dir, err := ioutil.TempDir("", "bob-test-aggregate-*")
	assert.Nil(t, err)

	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	assert.Nil(t, err)

	testBob, err := Bob(WithDir(dir), WithAllowRedundantTargets())
	assert.Nil(t, err)

	projectName := "first-level"
	projectNameSecondLevel := "second-level"
	projectNameThirdLevel := "third-level"

	err = CreatePlayground(
		PlaygroundOptions{
			Dir: dir,

			ProjectName:            projectName,
			ProjectNameSecondLevel: projectNameSecondLevel,
			ProjectNameThirdLevel:  projectNameThirdLevel,
		},
	)
	assert.Nil(t, err)

	aggregate, err := testBob.Aggregate()
	assert.Nil(t, err)

	for _, bobFile := range aggregate.Bobfiles() {
		assert.Equal(t, bobFile.Project, projectName)
	}
}
