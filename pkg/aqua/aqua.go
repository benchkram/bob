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

// Constants used internally
const (
	AQUA_ROOT      = ".bob/.aqua"
	AQUA_VERSION   = "v0.13.0"
	AQUA_FILE_PATH = ".bob/.aqua/aqua.yaml"

	ENV_AQUA_GLOBAL_CONFIG = "AQUA_GLOBAL_CONFIG"
	ENV_AQUA_ROOT_DIR      = "AQUA_ROOT_DIR"
	ENV_PATH               = "PATH"
)

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

// EnvironmentVariables holds paths necessary for aqua runtime environment
type EnvionmentVariables struct {
	AquaRoot   string
	AquaBin    string
	AquaConfig string
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
	err = os.MkdirAll(AQUA_ROOT, os.ModePerm)
	errz.Fatal(err)

	// Setup envirionment
	err = d.SetEnvirionment()
	errz.Fatal(err)

	// Create aqua.yaml file used inside aqua controller install call
	err = d.createDefinitionFile()
	errz.Fatal(err)

	param := &controller.Param{
		ConfigFilePath: AQUA_FILE_PATH, // This should be nested somewhere inside .bob dir
		IsTest:         false,
		All:            true,
		AQUAVersion:    AQUA_VERSION, // we need to regularly check and update this version
	}

	ctrl, err := controller.New(ctx, param)
	errz.Fatal(err)

	// Finally call installation
	err = ctrl.Install(context.Background(), param)
	errz.Fatal(err)

	return nil
}

// Setup run/build environment so that packages can be accessed
func (d *Definition) SetEnvirionment() error {
	// Will hide internally used paths
	_, err := d.setEnvirionment()
	return err
}

func (d *Definition) setEnvirionment() (env EnvionmentVariables, err error) {
	defer errz.Recover(&err)

	// Add aqua root to env
	aquaRoot, err := filepath.Abs(AQUA_ROOT)
	errz.Fatal(err)

	err = os.Setenv(ENV_AQUA_ROOT_DIR, aquaRoot)
	errz.Fatal(err)

	// Add global aqua config file to env
	aquaConfig, err := filepath.Abs(AQUA_FILE_PATH)
	errz.Fatal(err)

	err = os.Setenv(ENV_AQUA_GLOBAL_CONFIG, aquaConfig)

	// Add aqua bin to PATH
	aquaBin, err := filepath.Abs(fmt.Sprintf("%s/bin", aquaRoot))
	errz.Fatal(err)

	err = os.Setenv(ENV_PATH, fmt.Sprintf("%s:%s", aquaBin, os.Getenv(ENV_PATH)))
	errz.Fatal(err)

	env = EnvionmentVariables{
		AquaRoot:   aquaRoot,
		AquaBin:    aquaBin,
		AquaConfig: aquaConfig,
	}
	return env, err
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
