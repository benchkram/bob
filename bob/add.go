package bob

import (
	"net/url"
	"strings"

	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
)

// Add, adds the rawurl as a repositoroy to bob workspace. Set all url if
// explicit url set to false.
//
// if explicit protocol set to true,
//
// set the sshurl & localurl to "" if url is a valid http,
//
// set the httpurl and sshurl to "" if url is a filepath
//
// else it presumes url type as ssh and set http url to ""
func (b *B) Add(rawurl string, explcitprotcl bool) (err error) {
	defer errz.Recover(&err)

	// check if url ends with `.git`
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

	if explcitprotcl && checkIfHttp(rawurl) {
		sshstr = ""
		localstr = ""
	} else if explcitprotcl && checkIfFile(rawurl) {
		httpsstr = ""
		sshstr = ""
	} else if explcitprotcl {
		httpsstr = ""
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

// checkIfHttp returns true if url is http,
func checkIfHttp(rawurl string) bool {
	err := isValidUrl(rawurl)
	if err != nil {
		return false
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

// checkIfFile returns true if url is filepath
func checkIfFile(rawurl string) bool {
	err := isValidUrl(rawurl)
	if err != nil {
		return false
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}

	return u.Scheme == "file"
}

// isValidUrl tests a string to determine if it is a well-structured url or not.
func isValidUrl(toTest string) error {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return err
	}

	return nil
}
