package bobtask

import "fmt"

var (
	ErrBobfileNotFound        = fmt.Errorf("could not find a bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("task hash does not exist")
	ErrHashInDoesNotExist     = fmt.Errorf("input-hash does not exist")
	ErrInvalidInput           = fmt.Errorf("invalid input")
	ErrBuildinfostoreIsNil    = fmt.Errorf("buildinfostore is nil")

	ErrInvalidTargetDefinition  = fmt.Errorf("invalid target definition, can't find 'path' or 'image' directive")
	ErrAmbigousTargetDefinition = fmt.Errorf("ambigous target definition, can't have 'path' and 'image' directive on same target")
)
