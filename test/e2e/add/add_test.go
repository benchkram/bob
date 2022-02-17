package addtest

import (
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/add"
	giturls "github.com/whilp/git-urls"

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
				Expect(b.Add(fmt.Sprintf("file://%s", child), false)).NotTo(HaveOccurred())
			}
		})

		It("adds HTTPS repo to bob, with plain protocol", func() {
			targeturl := "https://github.com/pkg/requests.git"
			Expect(b.Add(targeturl, true)).NotTo(HaveOccurred())

			repo, err := getRepositoryByUrlFromBob(b, targeturl)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.SSHUrl).To(Equal(""))
			Expect(repo.LocalUrl).To(Equal(""))
			Expect(repo.HTTPSUrl).To(Equal(targeturl))
		})

		It("adds SSH repo to bob, with plain protocol", func() {
			targeturl := "git@github.com:pkg/exec.git"
			Expect(b.Add(targeturl, true)).NotTo(HaveOccurred())

			repo, err := getRepositoryByUrlFromBob(b, targeturl)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.HTTPSUrl).To(Equal(""))
			Expect(repo.LocalUrl).To(Equal(""))
			Expect(repo.SSHUrl).To(Equal(targeturl))
		})

		It("adds Empty url, must be failed", func() {
			Expect(b.Add("", true)).To(HaveOccurred())
		})

		It("adds Empty https url, does not fail with strings starts with valid protocol", func() {
			err := b.Add("https://", true)
			Expect(err).NotTo(HaveOccurred())
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

func getRepositoryByUrlFromBob(b *bob.B, url string) (*bob.Repo, error) {
	repourl, err := giturls.Parse(url)
	if err != nil {
		return nil, err
	}

	name := bob.RepoName(repourl)

	for _, r := range b.Repositories {
		if r.Name == name {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("url not found to be added on repositorues")

}
