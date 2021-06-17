package bob

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"

	"github.com/Benchkram/errz"
)

var azureGitProvider = GitProvider{Name: "azure", Domain: "dev.azure.com"}
var githubGitProvider = GitProvider{Name: "github", Domain: "github.com"}
var localGitProvider = GitProvider{Name: "file://"}

var parsers = NewParsers()

type GitProvider struct {
	Name   string
	Domain string

	Parse Parser
}

type GitRepo struct {
	Provider GitProvider

	HTTPS *GitURL
	SSH   *GitURL
	Local string
}

type Parsers map[string]Parser

func NewParsers() Parsers {
	providers := make(map[string]Parser)
	providers[azureGitProvider.Name] = ParseAzure
	providers[githubGitProvider.Name] = ParseGithub
	providers[localGitProvider.Name] = ParseLocal
	return providers
}

type Parser func(string) (*GitRepo, error)

// Parse a rawurl string and return a GitRepo object containing
// the specific https and ssh protocol urls.
func Parse(rawurl string) (repo *GitRepo, err error) {
	for _, parser := range parsers {
		repo, err := parser(rawurl)
		if err == nil {
			return repo, nil
		}
	}
	return nil, fmt.Errorf("Could not parse url to a known git providers")
}

// ParseAzure parses a it repo url and return the corresponding git & https protocol links
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

// ParseGithub parses a it repo url and return the corresponding git & https protocol links
// corresponding to azure-devops domain specifications.
// https://github.com/Benchkram/bob.git
// git@github.com:Benchkram/bob.git
func ParseGithub(rawurl string) (repo *GitRepo, err error) {
	defer errz.Recover(&err)
	if !strings.Contains(rawurl, githubGitProvider.Name) {
		return nil, fmt.Errorf("Could not parse %s as %s-repo", rawurl, githubGitProvider.Name)
	}

	repo = &GitRepo{
		Provider: githubGitProvider,
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

	return nil, fmt.Errorf("Could not detect a valid %s url", githubGitProvider.Name)
}

func ParseLocal(rawurl string) (repo *GitRepo, err error) {
	defer errz.Recover(&err)
	if !strings.Contains(rawurl, localGitProvider.Name) {
		return nil, fmt.Errorf("Could not parse %s as %s-repo", rawurl, localGitProvider.Name)
	}

	repo = &GitRepo{
		Provider: localGitProvider,
	}

	u, err := giturls.Parse(rawurl)
	errz.Fatal(err)

	repo.HTTPS = FromURL(u)
	repo.SSH = FromURL(u)
	repo.Local = strings.TrimPrefix(rawurl, "file://")

	return repo, nil
}

// func ParseGithub(rawurl string) (*url.URL, error) {

// 	return nil, nil
// }

// func ParseGitlab(rawurl string) (*url.URL, error) {

// 	return nil, nil
// }

// func ParseBitbucket(rawurl string) (*url.URL, error) {

// 	return nil, nil
// }

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
