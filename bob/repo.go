package bob

import (
	"net/url"
	"path"
	"strings"

	"github.com/benchkram/errz"
)

type Repo struct {
	Name string

	SSHUrl   string
	HTTPSUrl string
	LocalUrl string
}

// func newRepo() *Repo {
// 	r := &Repo{}
// 	return r
// }

func (b *B) RepositoryNames() (names []string, err error) {
	defer errz.Recover(&err)

	names = []string{}
	for _, repo := range b.Repositories {
		names = append(names, repo.Name)
	}

	return names, nil
}

// func (b *B) repos() (urls []*url.URL, err error) {
// 	defer errz.Recover(&err)

// 	urls = []*url.URL{}
// 	for _, repo := range c.Repositories {
// 		url, err := giturls.Parse(repo)
// 		errz.Fatal(err)
// 		urls = append(urls, url)
// 	}

// 	return urls, nil
// }

// RepoName returns the base part of the repo
// as the name. Suffix excluded.
func RepoName(repoURL *url.URL) (name string) {

	base := path.Base(repoURL.String())
	ext := path.Ext(repoURL.String())

	name = base
	if ext == ".git" {
		name = strings.TrimSuffix(base, ext)
	}

	return name
}
