package bob

import "fmt"

var (
	ErrConfigFileDoesNotExist      = fmt.Errorf("config file does not exist")
	ErrRepoAlreadyAdded            = fmt.Errorf("repo already added")
	ErrTaskDoesNotExist            = fmt.Errorf("task does not exist")
	ErrRunDoesNotExist             = fmt.Errorf("run does not exist")
	ErrWorkspaceAlreadyInitialised = fmt.Errorf("bob Workspace Already Initialized")
	ErrTargetValidationFailed      = fmt.Errorf("target validation failed")
	ErrCouldNotFindTopLevelBobfile = fmt.Errorf("could not find top-level Bobfile")
	ErrInvalidVersion              = fmt.Errorf("invalid version")
	ErrInsecuredHTTPURL            = fmt.Errorf("insecured http url not supported")
	ErrInvalidScheme               = fmt.Errorf("invalid scheme")
	ErrInvalidGitUrl               = fmt.Errorf("invalid git url")
)
