package bobtask

import "fmt"

var (
	ErrBobfileNotFound        = fmt.Errorf("could not find a bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("task hash does not exist")
	ErrHashInDoesNotExist     = fmt.Errorf("input-hash does not exist")
	ErrTaskDoesNotExist       = fmt.Errorf("task does not exist")
	ErrInvalidInput           = fmt.Errorf("invalid input")
)
