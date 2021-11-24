package reposetup

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	dirMode  = 0755
	fileMode = 0644

	fileData = "hello world"
)

/*
Directory structure:

files
├── deeplynested
│   └── a
│       ├── b
│       │   ├── c
│       │   │   ├── d
│       │   │   │   ├── f_d_one
│       │   │   │   ├── f_d_two
│       │   │   │   ├── f_one
│       │   │   │   ├── f_three
│       │   │   │   └── f_two
│       │   │   ├── f_c_one
│       │   │   └── f_c_two
│       │   ├── f_b_one
│       │   └── f_b_two
│       └── f_a_one
├── hardlinks
├── regulars
│   ├── f_one
│   ├── f_three
│   └── f_two
├── symlinks
└── wildcards
    └── a
        ├── b
        │   ├── c
        │   │   ├── d
        │   │   │   ├── f_1
        │   │   │   ├── f_2
        │   │   │   ├── f_3
        │   │   │   ├── f_d_1
        │   │   │   └── f_d_2
        │   │   ├── f_c_1
        │   │   └── f_c_2
        │   ├── f_b_1
        │   └── f_b_2
        └── f_a_1
*/
func SetupBaseFileStructure(basePath string) []string {
	dir := filepath.Join(basePath, "files")
	dir = mkdir(dir)

	var files []string

	files = append(files, setupRegularFiles(dir)...)
	files = append(files, setupDeeplyNestedFiles(dir)...)
	files = append(files, setupWildcardFiles(dir)...)
	files = append(files, setupSymlinkedFiles(dir)...)
	files = append(files, setupHardlinkedFiles(dir)...)

	return files
}

func mkdir(path ...string) string {
	dir := filepath.Join(path...)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		panic(err)
	}

	return dir
}

func mkfile(path ...string) string {
	file := filepath.Join(path...)
	if err := ioutil.WriteFile(file, []byte(fileData), fileMode); err != nil {
		panic(err)
	}

	return file
}

// Set up regular files.
func setupRegularFiles(basePath string) []string {
	dir := filepath.Join(basePath, "regulars")
	dir = mkdir(dir)

	var files []string

	files = append(files, mkfile(dir, "f_one"))
	files = append(files, mkfile(dir, "f_two"))
	files = append(files, mkfile(dir, "f_three"))

	return files
}

// Set up deeply nested files.
func setupDeeplyNestedFiles(basePath string) []string {
	dir := filepath.Join(basePath, "deeplynested")
	dirA := filepath.Join(dir, "a")
	dirB := filepath.Join(dir, "a/b")
	dirC := filepath.Join(dir, "a/b/c")
	dirD := filepath.Join(dir, "a/b/c/d")
	dirLast := dirD
	dir = mkdir(dirLast)

	var files []string

	files = append(files, mkfile(dir, "f_one"))
	files = append(files, mkfile(dir, "f_two"))
	files = append(files, mkfile(dir, "f_three"))

	files = append(files, mkfile(dirA, "f_a_one"))

	files = append(files, mkfile(dirB, "f_b_one"))
	files = append(files, mkfile(dirB, "f_b_two"))

	files = append(files, mkfile(dirC, "f_c_one"))
	files = append(files, mkfile(dirC, "f_c_two"))

	files = append(files, mkfile(dirD, "f_d_one"))
	files = append(files, mkfile(dirD, "f_d_two"))

	return files
}

// Set up bath structure used to apply wildcards.
func setupWildcardFiles(basePath string) []string {
	dir := filepath.Join(basePath, "wildcards")
	dirA := filepath.Join(dir, "a")
	dirB := filepath.Join(dir, "a/b")
	dirC := filepath.Join(dir, "a/b/c")
	dirD := filepath.Join(dir, "a/b/c/d")
	dirLast := dirD
	dir = mkdir(dirLast)

	var files []string

	files = append(files, mkfile(dir, "f_1"))
	files = append(files, mkfile(dir, "f_2"))
	files = append(files, mkfile(dir, "f_3"))

	files = append(files, mkfile(dirA, "f_a_1"))

	files = append(files, mkfile(dirB, "f_b_1"))
	files = append(files, mkfile(dirB, "f_b_2"))

	files = append(files, mkfile(dirC, "f_c_1"))
	files = append(files, mkfile(dirC, "f_c_2"))

	files = append(files, mkfile(dirD, "f_d_1"))
	files = append(files, mkfile(dirD, "f_d_2"))

	return files
}

// Set up symlinked files.
func setupSymlinkedFiles(basePath string) []string {
	dir := filepath.Join(basePath, "symlinks")
	// dir = mkdir(dir)
	_ = mkdir(dir)

	return nil
}

// Set up hardlinked files.
func setupHardlinkedFiles(basePath string) []string {
	dir := filepath.Join(basePath, "hardlinks")
	// dir = mkdir(dir)
	_ = mkdir(dir)

	return nil
}
