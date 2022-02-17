package bob

import (
	"strings"

	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	giturls "github.com/whilp/git-urls"
)

// Add, adds the rawurl as a repositoroy to bob workspace. Set all url if
// `plain` variable set to false.
//
// if plain set to true,
//
// set the sshurl & localurl to "" if url is a valid http,
// set the httpurl and sshurl to "" if url is a filepath,
// else it presumes url type as ssh and set http url to "".
func (b *B) Add(rawurl string, plain bool) (err error) {
	defer errz.Recover(&err)

	// check if remote url ends with `.git`
	isValid := checkIfURLEndsWithGit(rawurl)
	if !isValid {
		return usererror.Wrapm(ErrInvalidURL, "GIT url Add failed")
	}

	// Check if it is a valid git repo
	repo, err := Parse(rawurl)
	errz.Fatal(err)

	// TODO: let repoName be handled by Parse().
	name := RepoName(repo.HTTPS.URL)

	// Check for duplicates
	for _, existingRepo := range b.Repositories {
		if existingRepo.Name == name {
			return usererror.Wrapm(ErrRepoAlreadyAdded, "GIT url Add failed")
		}
	}

	httpsstr := repo.HTTPS.String()
	sshstr := repo.SSH.String()
	localstr := repo.Local

	if plain {
		scheme, err := getScheme(rawurl)
		errz.Fatal(err)

		switch scheme {
		case "http":
			return usererror.Wrapm(ErrInsecuredHTTPURL, "GIT url Add failed")
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
			// should not do anything
		}

	}

	b.Repositories = append(b.Repositories,
		Repo{
			Name:     name,
			HTTPSUrl: httpsstr,
			SSHUrl:   sshstr,
			LocalUrl: localstr,
		},
	)

	err = b.gitignoreAdd(name)
	errz.Fatal(err)

	return b.write()
}

// checkIfURLEndsWithGit checks if the provided url string
// has `.git` or `.git/` suffix on its end.
//
// Ignores local urls.
func checkIfURLEndsWithGit(rawurl string) bool {
	if checkIfFile(rawurl) {
		return true
	}

	trimmed := strings.TrimSuffix(rawurl, "/")
	if strings.HasSuffix(trimmed, ".git") {
		return true
	} else {
		return false
	}
}

// checkIfFile returns true if url is filepath
func checkIfFile(rawurl string) bool {
	scheme, err := getScheme(rawurl)
	if err != nil {
		return false
	}

	return scheme == "file"
}

// getScheme check if the url is valid,
// if valid returns the url scheme
func getScheme(rawurl string) (string, error) {
	u, err := giturls.Parse(rawurl)
	if err != nil {
		return "", err
	}

	if strings.Contains(rawurl, "git@") {
		return "ssh", nil
	}

	return u.Scheme, nil
}
