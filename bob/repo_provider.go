package bob

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"

	"github.com/benchkram/errz"
)

var generalGitProvider = GitProvider{Name: "general"}

// special providers with non trival to parse urls
var azureGitProvider = GitProvider{Name: "azure", Domain: "dev.azure.com"}
var localGitProvider = GitProvider{Name: "file://"}

var parsers = NewParsers()

type GitProvider struct {
	Name   string
	Domain string

	Parse Parser
}

type GitRepo struct {
	Provider GitProvider

	// SSH stores git@
	SSH *GitURL
	// HTTPS stores https://
	HTTPS *GitURL
	// Local stores file://
	Local string
}

// Name of the repository, usually part
// after the last "/".
func (gr *GitRepo) Name() string {

	if gr.SSH != nil && gr.SSH.URL != nil {
		return Name(gr.SSH.URL.String())
	}
	if gr.HTTPS != nil && gr.HTTPS.URL != nil {
		return Name(gr.HTTPS.URL.String())
	}
	if gr.Local != "" {
		return Name(gr.Local)
	}

	return ""

}

// RepoName returns the base part of the repo
// as the name. Suffix excluded.
func Name(repoURL string) (name string) {

	base := path.Base(repoURL)
	ext := path.Ext(repoURL)

	name = base
	if ext == ".git" {
		name = strings.TrimSuffix(base, ext)
	}

	return name
}

type Parsers map[string]Parser

func NewParsers() Parsers {
	providers := make(map[string]Parser)
	providers[azureGitProvider.Name] = ParseAzure
	providers[localGitProvider.Name] = ParseLocal
	return providers
}

type Parser func(string) (*GitRepo, error)

// Parse a rawurl and return a GitRepo object containing
// the specific https and ssh protocol urls.
func Parse(rawurl string) (repo *GitRepo, err error) {
	for _, parser := range parsers {
		repo, err := parser(rawurl)
		if err == nil {
			return repo, nil
		}
	}
	return ParseGeneral(rawurl)
}

// ParseAzure parses a git repo url and return the corresponding git & https protocol links
// corresponding to azure-devops domain specifications.
// https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz",
// git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz",
func ParseAzure(rawurl string) (repo *GitRepo, err error) {
	defer errz.Recover(&err)

	if !strings.Contains(rawurl, azureGitProvider.Name) {
		return nil, fmt.Errorf("Could not parse %s as %s-repo", rawurl, azureGitProvider.Name)
	}

	repo = &GitRepo{
		Provider: azureGitProvider,
	}

	u, err := giturls.Parse(rawurl)
	errz.Fatal(err)

	// Construct git url from https rawurl
	if strings.Contains(rawurl, "https") && strings.Contains(rawurl, "_git") {
		// Detected a http input url
		path := strings.Replace(u.Path, "/_git/", "/", 1)
		ssh := &url.URL{
			Host: "ssh." + u.Host + ":v3",
			Path: path,
			User: url.User("git"),
		}

		repo.HTTPS = FromURL(u)
		repo.SSH = FromURL(ssh)

		return repo, nil
	}

	// Construct https url from git rawurl
	if strings.Contains(rawurl, "git@") {
		// Detected a http input url
		host := strings.TrimPrefix(u.Host, "ssh.")

		// Trim `v3` and add `_git` in path
		// /v3/xxx/xxx/base => /xxx/xxx/_git/base
		path := strings.TrimPrefix(u.Path, "v3")
		base := filepath.Base(path)
		path = strings.TrimSuffix(path, base)
		path = filepath.Join(path, "_git", base)

		// Username is the first part of path.
		// Path starts with `/` => username is on splits[1].
		splits := strings.Split(path, "/")
		var user string
		if len(splits) > 1 {
			user = splits[1]
		}

		https := &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
			User:   url.User(user),
		}

		// Adjust ssh url
		u.Scheme = ""
		u.Path = strings.TrimPrefix(u.Path, "v3")
		u.Host = u.Host + ":v3"

		repo.HTTPS = FromURL(https)
		repo.SSH = FromURL(u)

		return repo, nil
	}

	return nil, fmt.Errorf("Could not detect a valid %s url", azureGitProvider.Name)
}

// ParseGeneral parses a git repo url and return the corresponding git & https protocol links
// corresponding to most (github, gitlab) domain specifications.
//
// github
// https://github.com/benchkram/bob.git
// git@github.com:benchkram/bob.git
// gitlab
// git@gitlab.com:gitlab-org/gitlab.git
// https://gitlab.com/gitlab-org/gitlab.git
//
func ParseGeneral(rawurl string) (repo *GitRepo, err error) {
	defer errz.Recover(&err)
	repo = &GitRepo{
		Provider: generalGitProvider,
	}

	u, err := giturls.Parse(rawurl)
	errz.Fatal(err)

	// Construct git url from https rawurl
	if strings.Contains(rawurl, "https") {
		// Username is the first part of path.
		// Path starts with `/` => username is on splits[1].
		splits := strings.Split(u.Path, "/")
		var user string
		if len(splits) > 1 {
			user = splits[1]
		}

		host := u.Host + ":" + user
		path := strings.TrimPrefix(u.Path, "/"+user)

		ssh := &url.URL{
			Host: host,
			Path: path,
			User: url.User("git"),
		}

		repo.HTTPS = FromURL(u)
		repo.SSH = FromURL(ssh)

		return repo, nil
	}

	// Construct https url from git rawurl
	if strings.Contains(rawurl, "git@") {
		https := &url.URL{
			Scheme: "https",
			Host:   u.Host,
			Path:   u.Path,
		}

		// Username is the first part of path.
		splits := strings.Split(u.Path, "/")
		var user string
		if len(splits) > 0 {
			user = splits[0]
		}

		host := u.Host + ":" + user
		path := strings.TrimPrefix(u.Path, user)

		// Adjust ssh url
		u.Scheme = ""
		u.Host = host
		u.Path = path

		repo.HTTPS = FromURL(https)
		repo.SSH = FromURL(u)

		return repo, nil
	}

	return nil, ErrInvalidGitUrl
}

func ParseLocal(rawurl string) (repo *GitRepo, err error) {
	defer errz.Recover(&err)
	if !strings.Contains(rawurl, localGitProvider.Name) {
		return nil, fmt.Errorf("Could not parse %s as %s-repo", rawurl, localGitProvider.Name)
	}

	repo = &GitRepo{
		Provider: localGitProvider,
	}

	// u, err := giturls.Parse(rawurl)
	// errz.Fatal(err)

	// repo.HTTPS = FromURL(u)
	// repo.SSH = FromURL(u)
	repo.Local = strings.TrimPrefix(rawurl, "file://")

	return repo, nil
}

// GitURL overlays `url.URL` to handle git clone urls correctly.
type GitURL struct {
	URL *url.URL
}

func FromURL(u *url.URL) *GitURL {
	return &GitURL{URL: u}
}
func (g *GitURL) String() string {
	if g.URL.Scheme == "" {
		return strings.TrimPrefix(g.URL.String(), "//")
	}
	return g.URL.String()
}
