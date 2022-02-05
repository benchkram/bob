package aqua

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	AQUA_PACKAGE_PREFIX = "- name: "
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
			Name: fmt.Sprintf("%s@%s", p.Name, p.Version),
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
	aquaEnv, err := d.setEnvirionment()
	errz.Fatal(err)

	// Create aqua.yaml file used inside aqua controller install call
	err = d.createDefinitionFile(aquaEnv.AquaConfig)
	errz.Fatal(err)

	param := &controller.Param{
		ConfigFilePath: aquaEnv.AquaConfig, // This should be nested somewhere inside .bob dir
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

// Prune - remove all packages and undelying structures
func (d *Definition) Prune(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	aquaRoot, err := filepath.Abs(AQUA_ROOT)
	errz.Fatal(err)

	err = os.RemoveAll(aquaRoot)
	errz.Fatal(err)

	return nil
}

// Setup run/build environment so that packages can be accessed
func (d *Definition) SetEnvirionment() error {
	// Will hide internally used paths
	_, err := d.setEnvirionment()
	return err
}

func (d *Definition) Search(ctx context.Context) (pckgs []string, err error) {
	defer errz.Recover(&err)
	// Create local .aqua dir and reference to that
	err = os.MkdirAll(AQUA_ROOT, os.ModePerm)
	errz.Fatal(err)

	// Setup envirionment
	aquaEnv, err := d.setEnvirionment()
	errz.Fatal(err)

	// Create aqua.yaml file used inside aqua controller install call
	err = d.createDefinitionFile(aquaEnv.AquaConfig)
	errz.Fatal(err)

	param := &controller.Param{
		ConfigFilePath: aquaEnv.AquaConfig, // This should be nested somewhere inside .bob dir
		IsTest:         false,
		All:            true,
		AQUAVersion:    AQUA_VERSION, // we need to regularly check and update this version
	}

	ctrl, err := controller.New(ctx, param)

	errz.Fatal(err)

	// store old stdout
	old := ctrl.Stdout

	r, w, err := os.Pipe()
	errz.Fatal(err)

	ctrl.Stdout = w

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	err = ctrl.Generate(ctx, param)
	errz.Fatal(err)

	// back to normal state
	err = w.Close()
	errz.Fatal(err)
	ctrl.Stdout = old // restoring the real stdout
	out := <-outC

	// reading our temp stdout
	pckgs = strings.Split(out, "\n")

	// remove AQUA_PACKAGE_PREFIX to give a clean package name and version string to the caller
	for idx := range pckgs {
		pckgs[idx] = strings.TrimPrefix(pckgs[idx], AQUA_PACKAGE_PREFIX)
	}

	return pckgs, nil
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
func (d *Definition) createDefinitionFile(path string) (err error) {
	defer errz.Recover(&err)

	out, err := yaml.Marshal(d)
	errz.Fatal(err)

	err = os.WriteFile(path, out, os.ModePerm)
	errz.Fatal(err)

	return nil
}
