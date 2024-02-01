package collection

import (
	"fmt"
	"github.com/benchkram/bob/pkg/store-client/generated"
	"github.com/benchkram/bob/pkg/versionedsync/file"
	"github.com/benchkram/errz"
	"strings"
)

const (
	Divider = ";;;;;"
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
	return name + Divider + tag
}

func SplitName(combinedName string) (name, version string, _ error) {
	parts := strings.Split(combinedName, Divider)
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
			files = append(files, file.FromRestStubType(f))
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

// Len is the number of elements in the collection.
func (c C) Len() int {
	return len(c.Files)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (c C) Less(i, j int) bool {
	return c.Files[i].LocalPath < c.Files[j].LocalPath
}

// Swap swaps the elements with indexes i and j.
func (c C) Swap(i, j int) {
	tmpF := c.Files[i]
	c.Files[i] = c.Files[j]
	c.Files[j] = tmpF
}
