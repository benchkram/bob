package bobfile

import (
	"github.com/benchkram/errz"
)

// Verify a bobfile before task runner.
func (b *Bobfile) Verify() error {
	return b.verifyBefore()
}

// VerifyBefore a bobfile before task runner.
func (b *Bobfile) VerifyBefore() error {
	return b.verifyBefore()
}

// VerifyAfter a bobfile after task runner.
func (b *Bobfile) VerifyAfter() error {
	return b.verifyAfter()
}

// verifyBefore verifies a Bobfile before Run() is called.
func (b *Bobfile) verifyBefore() (err error) {
	defer errz.Recover(&err)

	// err = b.BTasks.VerifyDuplicateTargets()
	// errz.Fatal(err)

	for _, task := range b.BTasks {
		err = task.VerifyBefore()
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
