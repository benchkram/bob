package bobtask

import "fmt"

var (
	ErrBobfileNotFound        = fmt.Errorf("Could not find a Bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("Hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("Task hash does not exist")
	ErrHashInDoesNotExist     = fmt.Errorf("HashIn does not exist")
	ErrTaskDoesNotExist       = fmt.Errorf("Task does not exist")
)
