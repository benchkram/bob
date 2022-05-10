package project

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/benchkram/bob/pkg/usererror"
	"github.com/pkg/errors"
)

// RestrictedProjectNameChars collects the characters allowed to represent a project.
const RestrictedProjectNameChars = `[a-zA-Z0-9/_.\-:]`

// RestrictedProjectNamePattern is a regular expression to validate projectnames.
var RestrictedProjectNamePattern = regexp.MustCompile(`^` + RestrictedProjectNameChars + `+$`)

// ProjectNameDoubleSlashPattern matches a string containing a double slash (useful to check for URL schema)
var ProjectNameDoubleSlashPattern = regexp.MustCompile(`//+`)

var (
	ErrProjectIsLocal  = fmt.Errorf("can't use Remote() with local project")
	ErrProjectIsRemote = fmt.Errorf("can't use Local() with remote project")

	ErrInvalidProjectName = fmt.Errorf("invalid project name")

	ProjectNameFormatHint = "project name should be in the form 'project' or 'registry.com/user/project'"
)

type T string

const (
	Local  T = "local"
	Remote T = "remote"
)

type Name string

func (n *Name) Type() T {
	t, _, _ := parse(*n)
	return t
}

func (n *Name) Local() (string, error) {
	t, l, _ := parse(*n)
	switch t {
	case Local:
		return string(l), nil
	case Remote:
		return "", ErrProjectIsLocal
	default:
		return string(l), nil
	}
}

func (n *Name) Remote() (*url.URL, error) {
	t, _, url := parse(*n)
	switch t {
	case Local:
		return nil, ErrProjectIsRemote
	case Remote:
		return url, nil
	default:
		return url, nil
	}
}

// Parse a projectname and validate it against `RestrictedProjectNamePattern`
func Parse(projectname string) (Name, error) {
	if !RestrictedProjectNamePattern.MatchString(projectname) {
		return "", usererror.Wrap(errors.WithMessage(ErrInvalidProjectName,
			"project name should be in the form 'project' or 'registry.com/user/project'",
		))
	}

	// test for double slash (do not allow prepended schema)
	if ProjectNameDoubleSlashPattern.MatchString(projectname) {
		return "", usererror.Wrap(errors.WithMessage(ErrInvalidProjectName, ProjectNameFormatHint))
	}

	return Name(projectname), nil
}

func parse(projectname Name) (T, Name, *url.URL) {
	n := string(projectname)
	if n == "" {
		return Local, "", nil
	}

	segs := strings.Split(n, "/")
	if len(segs) <= 1 {
		return Local, projectname, nil
	}

	url, err := url.Parse("https://" + n)
	if err != nil {
		return Local, projectname, nil
	}

	// in case o a relative path expect it to be local
	if url.Host == "" {
		return Local, projectname, nil
	}

	url.Scheme = "https"
	
	return Remote, "", url
}
