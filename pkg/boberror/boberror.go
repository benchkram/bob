package boberror

import "fmt"

var ErrTaskDoesNotExist = fmt.Errorf("task does not exist")

func ErrTaskDoesNotExistF(task string) error {
	return fmt.Errorf("[task: %s], %w", task, ErrTaskDoesNotExist)
}
