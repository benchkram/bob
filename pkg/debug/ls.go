package debug

import (
	"os"
)

// LS list files in dir and print to stderr
func LS(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		println(f.Name())
	}
}

// LSWD
func LSWD() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	println("--------" + wd + "--------")
	LS(wd)
	println("--------")
}
