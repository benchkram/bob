package clonetest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/bob/build"
	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/test/repo/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob clone", func() {
	Context("in a fresh environment", func() {
		It("initializes bob", func() {
			Expect(b.Init()).NotTo(HaveOccurred())
		})

		It("adds HTTPS repo to bob", func() {
			Expect(b.Add("https://github.com/pkg/errors.git")).NotTo(HaveOccurred())
		})

		// TODO: Reenable. Fails to clone on CI.
		/*It("adds SSH repo to bob", func() {
			Expect(b.Add("git@github.com:Benchkram/bob.git")).NotTo(HaveOccurred())
		})*/

		It("adds local repos to bob", func() {
			// Children
			for _, child := range childs {
				Expect(b.Add(fmt.Sprintf("file://%s", child))).NotTo(HaveOccurred())
			}

			// Recursive
			Expect(b.Add(fmt.Sprintf("file://%s", recursiveRepo))).NotTo(HaveOccurred())
		})

		It("runs bob clone", func() {
			Expect(b.Clone()).NotTo(HaveOccurred())

			// Children
			for _, child := range setup.Childs {
				Expect(file.Exists(filepath.Join(top, child))).To(BeTrue())
			}

			// Recursive
			Expect(file.Exists(filepath.Join(top, setup.ChildRecursive))).To(BeTrue())
			// Delete HTTPS repo afterwards as it is reused later
			Expect(os.RemoveAll(filepath.Join(top, "errors"))).NotTo(HaveOccurred())
			// Delete recursive repo afterwards as it is reused later
			Expect(os.RemoveAll(filepath.Join(top, setup.ChildRecursive))).NotTo(HaveOccurred())
		})

		It("runs bob clone to clone a bob repo", func() {
			_, err := b.CloneRepo(fmt.Sprintf("file://%s", playgroundRepo))
			Expect(err).NotTo(HaveOccurred())

			f := filepath.Join(playgroundRepo, "second-level", "third-level", build.BobFileName)
			Expect(file.Exists(f)).To(BeTrue(), fmt.Sprintf("%s doesn't exist", f))
		})

		It("runs bob clone to clone a bob repo recursively", func() {
			_, err := b.CloneRepo(fmt.Sprintf("file://%s", recursiveRepo))
			Expect(err).NotTo(HaveOccurred())

			childProjectGit := filepath.Join(top, setup.ChildRecursive, "errors", ".git")
			// Make sure to not directly rely on a file in the repo
			// just assure that it's cloned correctly.
			Expect(file.Exists(childProjectGit)).To(BeTrue())
		})
	})
})
