package bob

import "fmt"

var (
	ErrConfigFileDoesNotExist      = fmt.Errorf("Config file does not exist")
	ErrRepoAlreadyAdded            = fmt.Errorf("Repo already added")
	ErrTaskDoesNotExist            = fmt.Errorf("Task does not exist")
	ErrRunDoesNotExist             = fmt.Errorf("Run does not exist")
	ErrWorkspaceAlreadyInitialised = fmt.Errorf("Bob Workspace Already Initialized")
	ErrTargetValidationFailed      = fmt.Errorf("Target validation failed")
	ErrCouldNotFindTopLevelBobfile = fmt.Errorf("Could not find top-level Bobfile")
	ErrInvalidVersion              = fmt.Errorf("invalid version")
)
