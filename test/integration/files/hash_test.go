package filestest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Benchkram/bob/bob/build"
	"github.com/Benchkram/bob/pkg/filepathutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	hashesFile = "./hashes"
)

var _ = Describe("Test hash-related functions", func() {
	Context("in a fresh environment", func() {
		It("lists all files recursively", func() {
			list, err := filepathutil.ListRecursive(dir)
			Expect(err).NotTo(HaveOccurred())

			err = hasSameElements(files, list)
			Expect(err).NotTo(HaveOccurred())
		})

		It("hashes the files", func() {
			list, err := filepathutil.ListRecursive(dir)
			Expect(err).NotTo(HaveOccurred())

			fhs := build.HashFiles(list)
			stripFileHashBase(fhs)
			// data, err := json.MarshalIndent(fhs, "", "\t")
			// Expect(err).NotTo(HaveOccurred())
			// fmt.Println(string(data))

			fhsFixtureRaw, err := ioutil.ReadFile(hashesFile)
			Expect(err).NotTo(HaveOccurred())

			var fhsFixture []build.FileHash
			err = json.Unmarshal(fhsFixtureRaw, &fhsFixture)
			Expect(err).NotTo(HaveOccurred())

			err = build.FileHashesDiffer(fhsFixture, fhs)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func stripBase(path string) string {
	return strings.Replace(path, dir, ".", 1)
}

func stripFileHashBase(fhs []build.FileHash) {
	for i, fh := range fhs {
		fhs[i].Path = stripBase(fh.Path)
	}
}

func hasSameElements(ss1, ss2 []string) error {
	for _, s1 := range ss1 {
		var found bool
		for _, s2 := range ss2 {
			if s1 == s2 {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in second slice", s1)
		}
	}

	for _, s2 := range ss2 {
		var found bool
		for _, s1 := range ss1 {
			if s1 == s2 {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in first slice", s2)
		}
	}

	return nil
}
