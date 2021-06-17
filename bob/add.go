package bob

import (
	"github.com/Benchkram/errz"
)

func (b *B) Add(rawurl string) (err error) {
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

	b.Repositories = append(b.Repositories,
		Repo{
			Name:     name,
			HTTPSUrl: repo.HTTPS.String(),
			SSHUrl:   repo.SSH.String(),
			LocalUrl: repo.Local,
		},
	)

	err = b.gitignoreAdd(name)
	errz.Fatal(err)

	return b.write()
}
