package bob

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/pkg/file"
)

type B struct {

	// DefaultCloneSchema used for cloning repos, `ssh` | `https`
	DefaultCloneSchema CloneSchema

	// Repositories to track.
	Repositories []Repo `yaml:"repositories"`

	// dir is bob's working directory.
	dir string

	// readConfig some commands need a fully initialised bob.
	// When this is true a `.bob/config` file must exist,
	// usually done by calling `bob init`
	readConfig bool
}

func new() *B {
	wd, err := os.Getwd()
	if err != nil {
		errz.Fatal(err)
	}

	c := &B{
		DefaultCloneSchema: SSH,

		dir: wd,
	}
	return c
}

func Bob(opts ...Option) (*B, error) {
	bob := new()

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(bob)
	}

	if bob.readConfig {
		err := bob.read()
		if err != nil {
			return nil, err
		}
	}

	return bob, nil
}

func (b *B) Dir() string {
	return b.dir
}

func (b *B) write() (err error) {
	defer errz.Recover(&err)

	bin, err := yaml.Marshal(b)
	errz.Fatal(err)

	const mode = 0644
	return ioutil.WriteFile(b.ConfigFilePath(), bin, mode)
}

func (b *B) read() (err error) {
	if !file.Exists(b.ConfigFilePath()) {
		// Initialise with default values if it does not exist.
		err := b.write()
		errz.Fatal(err)
	}

	bin, err := ioutil.ReadFile(b.ConfigFilePath())
	errz.Fatal(err, "Failed to read config file")

	err = yaml.Unmarshal(bin, b)
	errz.Fatal(err)

	return nil
}
