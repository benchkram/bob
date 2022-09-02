package bobtask

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/bobtask/targettype"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"gopkg.in/yaml.v3"
)

const (
	pathSelector  string = "path"
	imageSelector string = "image"
)

// parseTargets parses target definitions from yaml.
//
// Example yaml input:
//
// target: folder/
//
// target: |-
//	folder/
//	folder1/folder/file
//
// target:
//   path: |-
//		folder/
//		folder1/folder/file
//
// target:
//	image: docker-image-name
//
// target:
//   image: |-
//		docker-image-name
//		docker-image2-name
//
func (t *Task) parseTargets() error {

	var filesystemEntries []string
	var dockerImages []string
	var err error

	switch td := t.TargetDirty.(type) {
	case string:
		filesystemEntries, err = parseTargetPath(td)
	case map[string]interface{}:
		targets, targetType, err := parseTargetMap(td)
		if err != nil {
			err = usererror.Wrapm(err, fmt.Sprintf("[task:%s]", t.name))
		}

		switch targetType {
		case targettype.Path:
			filesystemEntries = targets
		case targettype.Docker:
			dockerImages = targets
		}
	}

	if err != nil {
		return err
	}

	if len(filesystemEntries) > 0 || len(dockerImages) > 0 {
		t.target = target.New(
			target.WithFilesystemEntries(filesystemEntries),
			target.WithDockerImages(dockerImages),
			target.WithDir(t.dir),
		)
	}

	return nil
}

func parseTargetMap(tm map[string]interface{}) ([]string, targettype.T, error) {

	// check first if both directives are selected
	if keyExists(tm, pathSelector) && keyExists(tm, imageSelector) {
		return nil, targettype.Path, ErrAmbigousTargetDefinition
	}

	paths, ok := tm[pathSelector]
	if ok {
		targets, err := parseTargetPath(paths.(string))
		if err != nil {
			return nil, targettype.Path, err
		}

		return targets, targettype.Path, nil
	}

	images, ok := tm[imageSelector]
	if !ok {
		return nil, targettype.Path, ErrInvalidTargetDefinition
	}

	return parseTargetImage(images.(string)), targettype.Docker, nil
}

func parseTargetPath(p string) ([]string, error) {
	targets := []string{}
	if p == "" {
		return targets, nil
	}

	targetStr := fmt.Sprintf("%v", p)
	targetDirty := split(targetStr)

	for _, targetPath := range unique(targetDirty) {
		if strings.Contains(targetPath, "../") {
			return targets, fmt.Errorf("'../' not allowed in file path %q", targetPath)
		}

		targets = append(targets, targetPath)
	}

	return targets, nil
}

func parseTargetImage(p string) []string {
	if p == "" {
		return []string{}
	}

	targetStr := fmt.Sprintf("%v", p)
	targetDirty := split(targetStr)

	return unique(targetDirty)
}

func keyExists(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func (t *Task) UnmarshalYAML(value *yaml.Node) (err error) {
	defer errz.Recover(&err)

	var values struct {
		Lowercase []string `yaml:"dependson"`
		Camelcase []string `yaml:"dependsOn"`
	}

	err = value.Decode(&values)
	errz.Fatal(err)

	if len(values.Lowercase) > 0 && len(values.Camelcase) > 0 {
		errz.Fatal(errors.New("both `dependson` and `dependsOn` nodes detected near line " + strconv.Itoa(value.Line)))
	}

	dependsOn := make([]string, 0)
	if values.Lowercase != nil && len(values.Lowercase) > 0 {
		dependsOn = values.Lowercase
	}
	if values.Camelcase != nil && len(values.Camelcase) > 0 {
		dependsOn = values.Camelcase
	}

	// new type needed to avoid infinite loop
	type TmpTask Task
	var tmpTask TmpTask

	err = value.Decode(&tmpTask)
	errz.Fatal(err)

	tmpTask.DependsOn = dependsOn

	*t = Task(tmpTask)

	return nil
}
