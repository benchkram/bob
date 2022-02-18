package aqua

import (
	"bytes"
	"regexp"

	"github.com/Benchkram/errz"
	"github.com/cli/cli/git"
)

const AQUA_REMOTE = "git@github.com:aquaproj/aqua-registry.git"

// cache if latestRegistry has already been aquired during current runtime
var gotLatestRegistry = false

// getDefaultRegsitry - poll for latest registry version through git ls-remote
func getDefaultRegistry() (_ Registry, err error) {
	defer errz.Recover(&err)

	if gotLatestRegistry {
		return defaultRegistry, nil
	}

	cmd, err := git.GitCommand("ls-remote", "--tags", AQUA_REMOTE)
	errz.Fatal(err)

	var out bytes.Buffer
	cmd.Stdout = &out

	// Run git cmd and wait for completion.
	err = cmd.Run()
	errz.Fatal(err)

	// find latest version string in output
	versionReg := regexp.MustCompile(`\w*\s*refs\/tags\/(v?\d+.\d+.\d+)`)
	matches := versionReg.FindAllStringSubmatch(out.String(), -1)

	// No match found
	if len(matches) == 0 || len(matches[len(matches)-1]) != 2 {
		return defaultRegistry, nil
	}

	latest := matches[len(matches)-1][1]

	// Set latest registry
	defaultRegistry.Ref = latest
	gotLatestRegistry = true

	return defaultRegistry, nil
}
