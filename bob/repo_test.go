package bob

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	giturls "github.com/whilp/git-urls"
)

func TestRepoName(t *testing.T) {

	tests := []struct {
		input  string // url
		result string
	}{
		{
			input:  "git@ssh.dev.azure.com:v3/first-level/seconlevel/my.repo.name",
			result: "my.repo.name",
		},
		{
			input:  "git@github.com:bob/bob.git",
			result: "bob",
		},
	}

	for _, test := range tests {
		url, err := giturls.Parse(test.input)
		assert.Nil(t, err, test.input)
		assert.Equal(t, test.result, RepoName(url), test.input)
	}
}

func TestParseAzure(t *testing.T) {
	type input struct {
		rawurl string
	}
	type result struct {
		https string
		ssh   string
	}
	tests := []struct {
		name   string
		input  input
		result result
	}{
		{
			name: "azure-devops-https",
			input: input{
				rawurl: "https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz",
			},
			result: result{
				https: "https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz",
				ssh:   "git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz",
			},
		},
		{
			name: "azure-devops-ssh",
			input: input{
				rawurl: "git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz",
			},
			result: result{
				https: "https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz",
				ssh:   "git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz",
			},
		},
	}

	for _, test := range tests {
		fmt.Println(test.name)

		repo, err := ParseAzure(test.input.rawurl)
		assert.Nil(t, err, test.name)

		assert.Equal(t, test.result.https, repo.HTTPS.String())
		assert.Equal(t, test.result.ssh, repo.SSH.String())
	}
}

func TestParseGeneral(t *testing.T) {
	type input struct {
		rawurl string
	}
	type result struct {
		https string
		ssh   string
	}
	tests := []struct {
		name   string
		input  input
		result result
	}{
		{
			name: "github-https",
			input: input{
				rawurl: "https://github.com/Benchkram/bob.git",
			},
			result: result{
				https: "https://github.com/Benchkram/bob.git",
				ssh:   "git@github.com:Benchkram/bob.git",
			},
		},
		{
			name: "github-ssh",
			input: input{
				rawurl: "git@github.com:Benchkram/bob.git",
			},
			result: result{
				https: "https://github.com/Benchkram/bob.git",
				ssh:   "git@github.com:Benchkram/bob.git",
			},
		},
		{
			name: "subdomain-https",
			input: input{
				rawurl: "https://subdomain.gitlab.de/project/repo.git",
			},
			result: result{
				https: "https://subdomain.gitlab.de/project/repo.git",
				ssh:   "git@subdomain.gitlab.de:project/repo.git",
			},
		},
		{
			name: "subdomain-ssh",
			input: input{
				rawurl: "git@subdomain.gitlab.de:project/repo.git",
			},
			result: result{
				https: "https://subdomain.gitlab.de/project/repo.git",
				ssh:   "git@subdomain.gitlab.de:project/repo.git",
			},
		},
	}

	for _, test := range tests {
		fmt.Println(test.name)

		repo, err := ParseGeneral(test.input.rawurl)
		assert.Nil(t, err, test.name)

		assert.Equal(t, test.result.https, repo.HTTPS.String())
		assert.Equal(t, test.result.ssh, repo.SSH.String())
	}
}

// func TestRepoURL(t *testing.T) {
// 	// Gathering of http + ssh urls
// 	//
// 	// Github:
// 	//   https://github.com/Benchkram/bob.git
// 	//   git@github.com:Benchkram/bob.git
// 	// Azure:
// 	//   https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz
// 	//   git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz
// 	// Gitlab:
// 	//   https://gitlab.domain.de/xxx/zzz-zzz-zzz.git
// 	//   git@gitlab.domain.de:xxx/zzz-zzz-zzz.git
// 	// Bitbucket:
// 	//   https://equanox@bitbucket.org/equanox/build-server.git
// 	//   git@bitbucket.org:equanox/build-server.git

// 	type input struct {
// 		https string
// 		ssh   string
// 	}
// 	tests := []struct {
// 		name   string
// 		input  input
// 		result string
// 	}{
// 		{
// 			name: "github",
// 			input: input{
// 				https: "https://github.com/Benchkram/bob.git",
// 				ssh:   "git@github.com:Benchkram/bob.git",
// 			},
// 			result: "",
// 		},
// 		{
// 			name: "azure-devops",
// 			input: input{
// 				https: "https://xxx@dev.azure.com/xxx/Yyy/_git/zzz.zzz.zzz",
// 				ssh:   "git@ssh.dev.azure.com:v3/xxx/Yyy/zzz.zzz.zzz",
// 			},
// 			result: "",
// 		},
// 		// {
// 		// 	name: "gitlab",
// 		// 	input: input{
// 		// 		https: "https://gitlab.domain.de/xxx/zzz-zzz-zzz.git",
// 		// 		ssh:   "git@gitlab.domain.de:xxx/zzz-zzz-zzz.git",
// 		// 	},
// 		// 	result: "",
// 		// },
// 		// {
// 		// 	name: "bitbucket",
// 		// 	input: input{
// 		// 		https: "https://equanox@bitbucket.org/equanox/build-server.git",
// 		// 		ssh:   "git@bitbucket.org:equanox/build-server.git",
// 		// 	},
// 		// 	result: "",
// 		// },
// 	}

// 	for _, test := range tests {
// 		httpsURL, err := giturls.Parse(test.input.https)
// 		assert.Nil(t, err, test.input)
// 		fmt.Println(httpsURL.String())
// 		sshURL, err := giturls.Parse(test.input.ssh)
// 		assert.Nil(t, err, test.input)
// 		fmt.Println(sshURL.String())

// 	}
// }
