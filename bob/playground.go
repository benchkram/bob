package bob

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/bobrun"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/export"
	"github.com/benchkram/bob/pkg/cmdutil"
	"github.com/benchkram/bob/pkg/file"
)

const (
	BuildAllTargetName            = "all"
	BuildTargetwithdirsTargetName = "targetwithdirs"
	BuildAlwaysTargetName         = "always-build"

	BuildTargetDockerImageName     = "docker-image"
	BuildTargetDockerImagePlusName = "docker-image-plus"
	// BuildTargetBobTestImage intentionaly has a path separator
	// in the image name to assure temporary tar archive generation
	// works as intended (uses the image name as filename).
	BuildTargetBobTestImage     = "bob/testimage:latest"
	BuildTargetBobTestImagePlus = "bob/testimage/plus:latest"
)

func maingo(ver int) []byte {
	return []byte(fmt.Sprintf(`package main

import (
	"os"
	"os/signal"
)

func main() {
        println("Hello Playground v%d")

		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt)
		<-signalChannel
        println("Byebye Playground v%d")
}
`, ver, ver))
}

var gomod = []byte(`module example.com/m

go 1.16
`)

var openapi = []byte(`openapi: 3.0.3
info:
  version: 1.0.0
  title: Playground
  license:
    name: Benchkram Software GmbH

paths:
  /health:
    get:
      tags:
        - system
      operationId: health
      responses:
        200:
          description: OK
        503:
          description: Service Unavailable
`)

var openapiSecondLevel = []byte(`openapi: 3.0.3
info:
  version: 1.0.0
  title: Playground Second Level
  license:
    name: Benchkram Software GmbH

paths:
  /second/level/health:
    get:
      tags:
        - system
      operationId: health
      responses:
        200:
          description: OK
        503:
          description: Service Unavailable
`)

var dockerfileAlpine = []byte(`FROM alpine
`)

var dockerfileAlpinePlus = []byte(`FROM alpine
RUN touch file
`)

const SecondLevelDir = "second-level"
const SecondLevelOpenapiProviderDir = "openapi-provider-project"
const ThirdLevelDir = "third-level"

type PlaygroundOptions struct {
	Dir                    string
	ProjectName            string
	ProjectNameSecondLevel string
	ProjectNameThirdLevel  string
}

// CreatePlayground creates a default playground
// to test bob workflows. projectName is used in the top-level bobfile
func CreatePlayground(opts PlaygroundOptions) error {
	// TODO: check if dir is empty
	// TODO: empty dir after consent

	err := os.Chdir(opts.Dir)
	errz.Fatal(err)

	// first level
	err = ioutil.WriteFile("go.mod", gomod, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("main1.go", maingo(1), 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("openapi.yaml", openapi, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("docker-compose.yml", dockercompose, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("docker-compose.whoami.yml", dockercomposewhoami, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("Dockerfile", dockerfileAlpine, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile("Dockerfile.plus", dockerfileAlpinePlus, 0644)
	errz.Fatal(err)

	err = createPlaygroundBobfile(".", true, opts.ProjectName)
	errz.Fatal(err)

	b := newBob()
	err = b.Init()
	if err != nil {
		if !errors.Is(err, ErrWorkspaceAlreadyInitialised) {
			errz.Fatal(err)
		}
	}

	// Create Git repo
	err = ioutil.WriteFile(filepath.Join(b.dir, ".gitignore"), []byte(
		""+
			SecondLevelDir+"\n"+
			SecondLevelOpenapiProviderDir+"\n",
	), 0644)
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "init")
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "add", "-A")
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "commit", "-m", "Initial commit")
	errz.Fatal(err)

	// second level
	err = os.MkdirAll(SecondLevelDir, 0755)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(SecondLevelDir, "go.mod"), gomod, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(SecondLevelDir, "main2.go"), maingo(2), 0644)
	errz.Fatal(err)

	b = newBob()
	b.dir = filepath.Join(b.dir, SecondLevelDir)
	err = b.init()
	if err != nil {
		if !errors.Is(err, ErrWorkspaceAlreadyInitialised) {
			errz.Fatal(err)
		}
	}

	err = createPlaygroundBobfileSecondLevel(b.dir, true, opts.ProjectNameSecondLevel)
	errz.Fatal(err)

	err = ioutil.WriteFile(filepath.Join(SecondLevelDir, "openapi.yaml"), openapiSecondLevel, 0644)
	errz.Fatal(err)

	// Create Git repo
	err = ioutil.WriteFile(filepath.Join(b.dir, ".gitignore"), []byte(
		""+
			ThirdLevelDir+"\n",
	), 0644)
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "init")
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "add", "-A")
	errz.Fatal(err)
	err = cmdutil.RunGit(b.dir, "commit", "-m", "Initial commit")
	errz.Fatal(err)

	// second level - openapi-provider
	err = os.MkdirAll(SecondLevelOpenapiProviderDir, 0755)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(SecondLevelOpenapiProviderDir, "openapi.yaml"), openapi, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(SecondLevelOpenapiProviderDir, "openapi2.yaml"), openapi, 0644)
	errz.Fatal(err)
	err = createPlaygroundBobfileSecondLevelOpenapiProvider(SecondLevelOpenapiProviderDir, true)
	errz.Fatal(err)

	// Create Git repo
	err = cmdutil.RunGit(SecondLevelOpenapiProviderDir, "init")
	errz.Fatal(err)
	err = cmdutil.RunGit(SecondLevelOpenapiProviderDir, "add", "-A")
	errz.Fatal(err)
	err = cmdutil.RunGit(SecondLevelOpenapiProviderDir, "commit", "-m", "Initial commit")
	errz.Fatal(err)

	// third level
	thirdDir := filepath.Join(SecondLevelDir, ThirdLevelDir)
	err = os.MkdirAll(thirdDir, 0755)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(thirdDir, "go.mod"), gomod, 0644)
	errz.Fatal(err)
	err = ioutil.WriteFile(filepath.Join(thirdDir, "main3.go"), maingo(3), 0644)
	errz.Fatal(err)

	b3 := newBob()
	b3.dir = filepath.Join(b3.dir, thirdDir)
	err = b3.init()
	if err != nil {
		if !errors.Is(err, ErrWorkspaceAlreadyInitialised) {
			errz.Fatal(err)
		}
	}

	err = createPlaygroundBobfileThirdLevel(b3.dir, true, opts.ProjectNameThirdLevel)
	errz.Fatal(err)

	err = ioutil.WriteFile(filepath.Join(thirdDir, "openapi.yaml"), openapiSecondLevel, 0644)
	errz.Fatal(err)

	// Create Git repo
	err = cmdutil.RunGit(b3.dir, "init")
	errz.Fatal(err)
	err = cmdutil.RunGit(b3.dir, "add", "-A")
	errz.Fatal(err)
	err = cmdutil.RunGit(b3.dir, "commit", "-m", "Initial commit")
	errz.Fatal(err)

	return nil
}

func createPlaygroundBobfile(dir string, overwrite bool, projectName string) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()

	bobfile.Project = projectName

	bobfile.Imports = []string{
		"second-level",
		"openapi-provider-project",
	}

	bobfile.Variables["helloworld"] = "Hello World!"

	bobfile.BTasks[global.DefaultBuildTask] = bobtask.Task{
		InputDirty:   "./main1.go" + "\n" + "go.mod",
		CmdDirty:     "go build -o run",
		TargetDirty:  "run",
		RebuildDirty: string(bobtask.RebuildOnChange),
	}

	bobfile.BTasks[BuildAllTargetName] = bobtask.Task{
		InputDirty: "./main1.go",
		CmdDirty:   "go build -o run",
		DependsOn: []string{
			filepath.Join(SecondLevelDir, fmt.Sprintf("%s2", global.DefaultBuildTask)),
			filepath.Join(SecondLevelDir, ThirdLevelDir, "print"),
		},
		TargetDirty:  "run",
		RebuildDirty: string(bobtask.RebuildOnChange),
	}

	bobfile.BTasks[BuildAlwaysTargetName] = bobtask.Task{
		InputDirty:   "./main1.go" + "\n" + "go.mod",
		CmdDirty:     "go build -o run",
		TargetDirty:  "run",
		RebuildDirty: string(bobtask.RebuildAlways),
	}

	bobfile.BTasks["generate"] = bobtask.Task{
		InputDirty: "openapi.yaml",
		CmdDirty: strings.Join([]string{
			"mkdir -p rest-server/generated",
			"oapi-codegen -package generated -generate server \\\n\t${OPENAPI_PROVIDER_PROJECT_OPENAPI_OPENAPI} \\\n\t\t> rest-server/generated/server.gen.go",
			"oapi-codegen -package generated -generate types \\\n\t${OPENAPI_PROVIDER_PROJECT_OPENAPI_OPENAPI} \\\n\t\t> rest-server/generated/types.gen.go",
			"oapi-codegen -package generated -generate client \\\n\t${OPENAPI_PROVIDER_PROJECT_OPENAPI_OPENAPI} \\\n\t\t> rest-server/generated/client.gen.go",
		}, "\n"),
		DependsOn: []string{
			filepath.Join(SecondLevelOpenapiProviderDir, "openapi"),
		},
		TargetDirty: strings.Join([]string{
			"rest-server/generated/server.gen.go",
			"rest-server/generated/types.gen.go",
			"rest-server/generated/client.gen.go",
		}, "\n"),
	}

	bobfile.BTasks["slow"] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			"sleep 2",
			"touch slowdone",
		}, "\n"),
		TargetDirty: "slowdone",
	}

	// A run command to run a environment from a compose file
	bobfile.RTasks["environment"] = &bobrun.Run{
		Type: bobrun.RunTypeCompose,
	}

	bobfile.RTasks["whoami"] = &bobrun.Run{
		Type: bobrun.RunTypeCompose,
		Path: "docker-compose.whoami.yml",
		DependsOn: []string{
			"all",
			"environment",
		},
	}

	// A run command to run a binary
	bobfile.RTasks["binary"] = &bobrun.Run{
		Type: bobrun.RunTypeBinary,
		Path: "./run",
		DependsOn: []string{
			"all",
			"environment",
		},
	}

	bobfile.BTasks["print"] = bobtask.Task{
		CmdDirty: "echo ${HELLOWORLD}",
	}

	bobfile.BTasks["multilinetouch"] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			"mkdir -p \\\nmultilinetouch",
			"touch \\\n\tmultilinefile1 \\\n\tmultilinefile2 \\\n\t\tmultilinefile3 \\\n        multilinefile4",
			"touch \\\n  multilinefile5",
		}, "\n"),
	}

	bobfile.BTasks["ignoredInputs"] = bobtask.Task{
		InputDirty: "fileToWatch" + "\n" + "!fileToIgnore",
		CmdDirty:   "echo \"Hello from ignored inputs task\"",
	}

	bobfile.BTasks[BuildTargetwithdirsTargetName] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			"mkdir -p .bbuild/dirone/dirtwo",
			"touch .bbuild/dirone/fileone",
			"touch .bbuild/dirone/filetwo",
			"touch .bbuild/dirone/dirtwo/fileone",
			"touch .bbuild/dirone/dirtwo/filetwo",
		}, "\n"),
		TargetDirty: ".bbuild/dirone/",
	}

	m := make(map[string]interface{})
	m["image"] = BuildTargetBobTestImage
	bobfile.BTasks[BuildTargetDockerImageName] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			fmt.Sprintf("docker build -t %s .", BuildTargetBobTestImage),
		}, "\n"),
		TargetDirty: m,
	}

	m = make(map[string]interface{})
	m["image"] = BuildTargetBobTestImagePlus
	bobfile.BTasks[BuildTargetDockerImagePlusName] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			fmt.Sprintf("docker build -f Dockerfile.plus -t %s .", BuildTargetBobTestImagePlus),
		}, "\n"),
		TargetDirty: m,
	}

	return bobfile.BobfileSave(dir)
}

func createPlaygroundBobfileSecondLevelOpenapiProvider(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()

	exports := make(export.Map)
	exports["openapi"] = "openapi.yaml"
	exports["openapi2"] = "openapi2.yaml"
	bobfile.BTasks["openapi"] = bobtask.Task{
		Exports: exports,
	}
	return bobfile.BobfileSave(dir)
}

func createPlaygroundBobfileSecondLevel(dir string, overwrite bool, projectName string) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()
	bobfile.Version = "1.2.3"
	bobfile.Project = projectName

	bobfile.Imports = []string{"third-level"}

	bobfile.BTasks[fmt.Sprintf("%s2", global.DefaultBuildTask)] = bobtask.Task{
		InputDirty: "./main2.go",
		DependsOn: []string{
			filepath.Join(ThirdLevelDir, fmt.Sprintf("%s3", global.DefaultBuildTask)),
		},
		CmdDirty:    "go build -o runsecondlevel",
		TargetDirty: "runsecondlevel",
	}
	return bobfile.BobfileSave(dir)
}

func createPlaygroundBobfileThirdLevel(dir string, overwrite bool, projectName string) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()
	bobfile.Version = "4.5.6"
	bobfile.Project = projectName

	bobfile.BTasks[fmt.Sprintf("%s3", global.DefaultBuildTask)] = bobtask.Task{
		InputDirty:  "*",
		CmdDirty:    "go build -o runthirdlevel",
		TargetDirty: "runthirdlevel",
	}

	bobfile.BTasks["print"] = bobtask.Task{
		CmdDirty: "echo hello-third-level",
	}

	return bobfile.BobfileSave(dir)
}
