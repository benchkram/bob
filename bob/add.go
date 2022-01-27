package bob

import (
	"strings"

	"github.com/Benchkram/errz"
)

func (b *B) Add(rawurl string, explcitprotcl bool) (err error) {
	defer errz.Recover(&err)

	// Check if it is a valid git repo
	repo, err := Parse(rawurl)
	errz.Fatal(err)

	// TODO: let repoName be handled by Parse().
	name := RepoName(repo.HTTPS.URL)

	// Check for duplicates
	for _, existingRepo := range b.Repositories {
		if existingRepo.Name == name {
			return ErrRepoAlreadyAdded
		}
	}

	httpsstr := repo.HTTPS.String()
	sshstr := repo.SSH.String()

	// if explicit protocol set to true,
	// set the sshurl to "" if url starts with http,
	// else http url set to ""
	if explcitprotcl && checkIfHttp(rawurl) {
		sshstr = ""
	} else if explcitprotcl {
		httpsstr = ""
	}

	b.Repositories = append(b.Repositories,
		Repo{
			Name:     name,
			HTTPSUrl: httpsstr,
			SSHUrl:   sshstr,
			LocalUrl: repo.Local,
		},
	)

	err = b.gitignoreAdd(name)
	errz.Fatal(err)

	return b.write()
}

// checkIfHttp returns true if url is http,
//
// It is easier to detect if the url is http/s protocol.
func checkIfHttp(rawurl string) bool {
	return strings.HasPrefix(rawurl, "http")
}
