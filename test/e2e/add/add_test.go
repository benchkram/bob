package addtest

import (
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/add"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob add", func() {
	Context("in a fresh environment", func() {
		It("initializes bob", func() {
			Expect(b.Init()).NotTo(HaveOccurred())
		})

		It("adds HTTPS repo to bob", func() {
			Expect(b.Add("https://github.com/pkg/errors.git", false)).NotTo(HaveOccurred())
		})

		It("adds SSH repo to bob", func() {
			Expect(b.Add("git@github.com:Benchkram/bob.git", false)).NotTo(HaveOccurred())
		})

		It("adds local repos to bob", func() {
			for _, child := range childs {
				fmt.Println(child)
				Expect(b.Add(fmt.Sprintf("file://%s", child), false)).NotTo(HaveOccurred())
			}
		})

		It("adds HTTPS repo to bob, with plain protocol", func() {
			Expect(b.Add("https://github.com/pkg/requests.git", true)).NotTo(HaveOccurred())
		})

		It("adds SSH repo to bob, with plain protocol", func() {
			Expect(b.Add("git@github.com:pkg/exec.git", true)).NotTo(HaveOccurred())
		})

		It("adds Empty url, must be failed", func() {
			Expect(b.Add("", true)).To(HaveOccurred())
		})

		It("adds Empty https url, must be failed", func() {
			err := b.Add("https://", true)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, bob.ErrInvalidURL)).To(BeTrue())
		})

		It("adds Empty git url, must return Invalid URL error", func() {
			err := b.Add("git@", true)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, bob.ErrInvalidURL)).To(BeTrue())
		})

		It("Invalid https url without .git on its end, must be failed", func() {
			err := b.Add("https://github.com/pkg/browser", false)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, bob.ErrInvalidURL)).To(BeTrue())
		})

		It("verifies that adding a duplicate repo fails on a new bob instance", func() {
			owd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			err = os.Chdir(b.Dir())
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err := os.Chdir(owd)
				Expect(err).NotTo(HaveOccurred())
			}()

			err = add.Add("https://github.com/pkg/errors.git")
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, bob.ErrRepoAlreadyAdded)).To(BeTrue(), fmt.Sprintf("%v != %v", err, bob.ErrRepoAlreadyAdded))
		})
	})
})
