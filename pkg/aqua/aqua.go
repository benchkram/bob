package aqua

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/packagemanager"
	"github.com/Benchkram/errz"
	"github.com/aquaproj/aqua/pkg/controller"
	"gopkg.in/yaml.v3"
)

const AQUA_ROOT = ".bob/.aqua"
const AQUA_VERSION = "v0.13.0"
const AQUA_FILE_PATH = ".bob/.aqua/aqua.yaml"

// This is the standard aqua registry
var defaultRegistry = Registry{
	Type: "standard",
	Ref:  AQUA_VERSION,
}

// Definition of a aqua.yaml file, also implements packagemanager interface
type Definition struct {
	Registries []Registry `yaml:"registries"`
	Packages   []Package  `yaml:"packages"`
}

// Registry links aqua registries
type Registry struct {
	Type string `yaml:"type"`
	Ref  string `yaml:"ref"`
}

// Packages holds info about a package to be installed/managed by aqua
type Package struct {
	Name string `yaml:"name"`
}

// New aqua definition
func New() *Definition {
	return &Definition{
		Registries: []Registry{
			defaultRegistry,
		},
		Packages: []Package{},
	}
}

// Add packages which should be installed
func (d *Definition) Add(packages ...packagemanager.Package) {
	for _, p := range packages {
		d.Packages = append(d.Packages, Package{
			Name: fmt.Sprintf("%s@%s", p.Name, p.Version.String()),
		})
	}
}

// Install will install all referenced packages defined in the PackageManager and add them
// to the PATH environment variable
func (d *Definition) Install(ctx context.Context) (err error) {
	defer errz.Recover(&err)
	// Create local .aqua dir and reference to that
	os.MkdirAll(AQUA_ROOT, os.ModePerm)
	os.Setenv("AQUA_ROOT_DIR", AQUA_ROOT)

	aquabin, err := filepath.Abs(fmt.Sprintf("%s/bin", AQUA_ROOT))
	errz.Fatal(err)

	// Create aqua.yaml file used inside aqua controller install call
	err = d.createDefinitionFile()
	errz.Fatal(err)

	// Add aqua bin to PATH
	os.Setenv("PATH", fmt.Sprintf("%s:%s", aquabin, os.Getenv("PATH")))

	param := &controller.Param{
		// TODO: checkout contents of config file

		ConfigFilePath: AQUA_FILE_PATH, // This could be nested somewhere inside .bob dir
		IsTest:         false,
		All:            true,
		AQUAVersion:    AQUA_VERSION,
	}

	ctrl, err := controller.New(ctx, param)
	errz.Fatal(err)

	err = ctrl.Install(context.Background(), param)
	errz.Fatal(err)

	return nil
}

// createDefinitionFile as aqua.yaml holding information stored in Definition object
func (d *Definition) createDefinitionFile() (err error) {
	defer errz.Recover(&err)

	out, err := yaml.Marshal(d)
	errz.Fatal(err)

	err = os.WriteFile(AQUA_FILE_PATH, out, os.ModePerm)
	errz.Fatal(err)

	return nil
}
