package bob

import (
	"errors"
	"strings"

	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	giturls "github.com/whilp/git-urls"
)

// Add, adds a repositoroy to a workspace.
//
// Automatically tries to guess the contrary url (git@/https://)
// from the given rawurl. This behavior can be deactivated
// if plain is set to true. In case of a local path (file://)
// plain has no effect.
func (b *B) Add(rawurl string, plain bool) (err error) {
	defer errz.Recover(&err)

	// Check if it is a valid git repo
	repo, err := Parse(rawurl)
	if err != nil {
		if errors.Is(err, ErrInvalidGitUrl) {
			return usererror.Wrap(err)
		} else {
			errz.Fatal(err)
		}
	}

	// Check for duplicates
	for _, existingRepo := range b.Repositories {
		if existingRepo.Name == repo.Name() {
			return usererror.Wrapm(ErrRepoAlreadyAdded, "failed to add repository")
		}
	}

	var httpsstr string
	if repo.HTTPS != nil {
		httpsstr = repo.HTTPS.String()
	}
	var sshstr string
	if repo.SSH != nil {
		sshstr = repo.SSH.String()
	}
	localstr := repo.Local

	if plain {
		scheme, err := getScheme(rawurl)
		errz.Fatal(err)

		switch scheme {
		case "http":
			return usererror.Wrapm(ErrInsecuredHTTPURL, "failed to add repository")
		case "https":
			sshstr = ""
			localstr = ""
		case "file":
			httpsstr = ""
			sshstr = ""
		case "ssh":
			httpsstr = ""
			localstr = ""
		default:
			return usererror.Wrap(ErrInvalidScheme)
		}

	}

	b.Repositories = append(b.Repositories,
		Repo{
			Name:     repo.Name(),
			HTTPSUrl: httpsstr,
			SSHUrl:   sshstr,
			LocalUrl: localstr,
		},
	)

	err = b.gitignoreAdd(repo.Name())
	errz.Fatal(err)

	return b.write()
}

// getScheme check if the url is valid,
// if valid returns the url scheme
func getScheme(rawurl string) (string, error) {
	u, err := giturls.Parse(rawurl)
	if err != nil {
		return "", err
	}

	// giturl parse detects url like `git@github.com` without `:` as files
	// which is a wrong URL but logically a `ssh`
	if strings.Contains(rawurl, "git@") {
		return "ssh", nil
	}

	return u.Scheme, nil
}
