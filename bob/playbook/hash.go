package playbook

import "fmt"

// PreComputeInputHashes asynchronically
func (pb *Playbook) PreComputeInputHashes() (err error) {

	// 3 parallel goroutines was the fastest one
	// on a test project.
	const max = 3

	var errors []error
	sem := make(chan int, max)
	for _, t := range pb.Tasks {
		sem <- 1

		// compute input hash
		go func(tt *Status) {
			_, err := tt.HashIn()
			if err != nil {
				errors = append(errors, fmt.Errorf("error computing input hash on task %s, %w", tt.Name(), err))
			}
			<-sem
		}(t)

	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
