package playbook

import (
	"github.com/benchkram/errz"
)

// PreComputeInputHashes asynchronically
func (pb *Playbook) PreComputeInputHashes() error {

	// 3 parallel goroutines was the fastest one
	// on a test project.
	const max = 3

	sem := make(chan int, max)
	for _, t := range pb.Tasks {
		sem <- 1

		// input hash
		go func() {
			_, err := t.HashIn()
			errz.Log(err)
			<-sem
		}()

	}

	return nil
}
