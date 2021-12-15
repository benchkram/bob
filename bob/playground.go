package bob

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/bobrun"
	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/bobtask/export"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/file"
)

const (
	BuildAllTargetName    = "all"
	BuildAlwaysTargetName = "always-build"
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

const SecondLevelDir = "second-level"
const SecondLevelOpenapiProviderDir = "openapi-provider-project"
const ThirdLevelDir = "third-level"

// CreatePlayground creates a default playground
// to test bob workflows.
func CreatePlayground(dir string) error {
	// TODO: check if dir is empty
	// TODO: empty dir after consent

	err := os.Chdir(dir)
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

	err = createPlaygroundBobfile(".", true)
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

	err = createPlaygroundBobfileSecondLevel(b.dir, true)
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

	err = createPlaygroundBobfileThirdLevel(b3.dir, true)
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

func createPlaygroundBobfile(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()

	bobfile.Variables["helloworld"] = "Hello World!"

	bobfile.Tasks[global.DefaultBuildTask] = bobtask.Task{
		InputDirty:  "./main1.go" + "\n" + "go.mod",
		CmdDirty:    "go build -o run",
		TargetDirty: "run",
	}

	bobfile.Tasks[BuildAllTargetName] = bobtask.Task{
		InputDirty: "./main1.go",
		CmdDirty:   "go build -o run",
		DependsOn: []string{
			filepath.Join(SecondLevelDir, fmt.Sprintf("%s2", global.DefaultBuildTask)),
			filepath.Join(SecondLevelDir, ThirdLevelDir, "print"),
		},
		TargetDirty: "run",
	}

	bobfile.Tasks[BuildAlwaysTargetName] = bobtask.Task{
		InputDirty:   "./main1.go" + "\n" + "go.mod",
		CmdDirty:     "go build -o run",
		TargetDirty:  "run",
		RebuildDirty: string(bobtask.RebuildAlways),
	}

	bobfile.Tasks["generate"] = bobtask.Task{
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

	bobfile.Tasks["slow"] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			"sleep 2",
			"touch slowdone",
		}, "\n"),
		TargetDirty: "slowdone",
	}

	// A run command to run a environment from a compose file
	bobfile.Runs["environment"] = &bobrun.Run{
		Type: bobrun.RunTypeCompose,
	}

	bobfile.Runs["whoami"] = &bobrun.Run{
		Type: bobrun.RunTypeCompose,
		Path: "docker-compose.whoami.yml",
		DependsOn: []string{
			"all",
			"environment",
		},
	}

	// A run command to run a binary
	bobfile.Runs["binary"] = &bobrun.Run{
		Type: bobrun.RunTypeBinary,
		Path: "./run",
		DependsOn: []string{
			"all",
			"environment",
		},
	}

	bobfile.Tasks["print"] = bobtask.Task{
		CmdDirty: "echo ${HELLOWORLD}",
	}

	bobfile.Tasks["multilinetouch"] = bobtask.Task{
		CmdDirty: strings.Join([]string{
			"mkdir -p \\\nmultilinetouch",
			"touch \\\n\tmultilinefile1 \\\n\tmultilinefile2 \\\n\t\tmultilinefile3 \\\n        multilinefile4",
			"touch \\\n  multilinefile5",
		}, "\n"),
	}

	bobfile.Tasks["ignoredInputs"] = bobtask.Task{
		InputDirty: "fileToWatch" + "\n" + "!fileToIgnore",
		CmdDirty:   "echo \"Hello from ignored inputs task\"",
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
	bobfile.Tasks["openapi"] = bobtask.Task{
		Exports: exports,
	}
	return bobfile.BobfileSave(dir)
}

func createPlaygroundBobfileSecondLevel(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()
	bobfile.Version = "1.2.3"

	bobfile.Tasks[fmt.Sprintf("%s2", global.DefaultBuildTask)] = bobtask.Task{
		InputDirty: "./main2.go",
		DependsOn: []string{
			filepath.Join(ThirdLevelDir, fmt.Sprintf("%s3", global.DefaultBuildTask)),
		},
		CmdDirty:    "go build -o runsecondlevel",
		TargetDirty: "runsecondlevel",
	}
	return bobfile.BobfileSave(dir)
}

func createPlaygroundBobfileThirdLevel(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return bobfile.ErrBobfileExists
	}

	bobfile := bobfile.NewBobfile()
	bobfile.Version = "4.5.6"

	bobfile.Tasks[fmt.Sprintf("%s3", global.DefaultBuildTask)] = bobtask.Task{
		InputDirty:  "*",
		CmdDirty:    "go build -o runthirdlevel",
		TargetDirty: "runthirdlevel",
	}

	bobfile.Tasks["print"] = bobtask.Task{
		CmdDirty: "echo hello-third-level",
	}

	return bobfile.BobfileSave(dir)
}
