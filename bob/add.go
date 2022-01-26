package bob

import (
	"github.com/Benchkram/errz"
)

func (b *B) Add(rawurl string, httpsonly bool, sshonly bool) (err error) {
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
	if sshonly {
		httpsstr = ""
	}

	sshstr := repo.SSH.String()
	if httpsonly {
		sshstr = ""
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
