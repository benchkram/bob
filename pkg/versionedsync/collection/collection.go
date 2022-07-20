package collection

import (
	"fmt"
	"github.com/benchkram/bob/pkg/store-client/generated"
	"github.com/benchkram/bob/pkg/versionedsync/file"
	"github.com/benchkram/errz"
	"strings"
)

const (
	divider = ";;"
)

// C is the representation of a collection of synced files
type C struct {
	ID string

	// Name of the collection
	Name string
	// Version
	Version string
	// LocalPath is the path to the collections root folder
	LocalPath string
	// Files are a set of individual blobs belonging to this collection
	Files []*file.F
}

func JoinNameAndVersion(name, tag string) string {
	return name + divider + tag
}

func SplitName(combinedName string) (name, version string, _ error) {
	parts := strings.Split(combinedName, divider)
	if len(parts) == 1 {
		return combinedName, "", nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", fmt.Errorf("invalid combined collection name")
	}
}

func FromRestType(genC *generated.SyncCollection) (_ *C, err error) {
	defer errz.Recover(&err)
	var files []*file.F
	if genC.Files != nil {
		for _, f := range *genC.Files {
			files = append(files, file.FileFromRestStubType(f))
		}
	}
	name, version, err := SplitName(genC.Name)
	errz.Fatal(err)
	return &C{
		ID:        genC.Id,
		Name:      name,
		Version:   version,
		LocalPath: genC.LocalPath,
		Files:     files,
	}, nil
}

func (c *C) FileByPath(localPath string) (*file.F, bool) {
	if c.Files == nil {
		return nil, false
	}
	for _, f := range c.Files {
		if f.LocalPath == localPath {
			return f, true
		}
	}
	return nil, false
}
