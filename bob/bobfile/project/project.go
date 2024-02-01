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
	ErrProjectIsRemote = fmt.Errorf("can't use Local() with remote project")

	ErrInvalidProjectName = fmt.Errorf("invalid project name")

	ProjectNameFormatHint = "project name should be in the form 'project' or 'bob.build/user/project'"
)

type T string

const (
	Local  T = "local"
	Remote T = "remote"
)

type Name string

func (n *Name) Type(allowInsecure bool) T {
	t, _, _ := parse(*n, allowInsecure)
	return t
}

func (n *Name) Remote(allowInsecure bool) (*url.URL, error) {
	t, _, url := parse(*n, allowInsecure)
	switch t {
	case Local:
		return nil, ErrProjectIsRemote
	case Remote:
		return url, nil
	default:
		return url, nil
	}
}

// Parse a projectName and validate it against `RestrictedProjectNamePattern`
func Parse(projectName string) (Name, error) {
	if !RestrictedProjectNamePattern.MatchString(projectName) {
		return "", usererror.Wrap(errors.WithMessage(ErrInvalidProjectName,
			"project name should be in the form 'project' or 'bob.build/user/project'",
		))
	}

	// test for double slash (do not allow prepended schema)
	if ProjectNameDoubleSlashPattern.MatchString(projectName) {
		return "", usererror.Wrap(errors.WithMessage(ErrInvalidProjectName, ProjectNameFormatHint))
	}

	return Name(projectName), nil
}

func parse(projectName Name, allowInsecure bool) (T, Name, *url.URL) {
	n := string(projectName)
	if n == "" {
		return Local, "", nil
	}

	segs := strings.Split(n, "/")
	if len(segs) <= 1 {
		return Local, projectName, nil
	}

	url, err := url.Parse("https://" + n)
	if err != nil {
		return Local, projectName, nil
	}

	// in case o a relative path expect it to be local
	if url.Host == "" {
		return Local, projectName, nil
	}

	url.Scheme = scheme(allowInsecure)

	return Remote, "", url
}

func scheme(allowInsecure bool) string {
	if allowInsecure {
		return "http"
	}
	return "https"
}
