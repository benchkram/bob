package bobfile

import (
	"github.com/benchkram/errz"
)

// Verify a bobfile before task runner.
func (b *Bobfile) Verify(cacheEnabled, allowRedundantTargets bool) error {
	return b.verifyBefore(cacheEnabled, allowRedundantTargets)
}

// VerifyAfter a bobfile after task runner.
func (b *Bobfile) VerifyAfter() error {
	return b.verifyAfter()
}

// verifyBefore verifies a Bobfile before Run() is called.
func (b *Bobfile) verifyBefore(cacheEnabled, allowRedundantTargets bool) (err error) {
	defer errz.Recover(&err)

	err = b.BTasks.VerifyDuplicateTargets()
	errz.Fatal(err)

	err = b.BTasks.VerifyMandatoryInputs()
	errz.Fatal(err)

	for _, task := range b.BTasks {
		err = task.VerifyBefore(cacheEnabled, allowRedundantTargets)
		errz.Fatal(err)
	}

	return nil
}

// verifyAfter verifies a Bobfile after Run() is called.
func (b *Bobfile) verifyAfter() (err error) {
	defer errz.Recover(&err)

	for _, task := range b.BTasks {
		err = task.VerifyAfter()
		errz.Fatal(err)
	}

	return nil
}
