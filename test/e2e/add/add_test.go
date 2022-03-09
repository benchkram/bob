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

		It("should add https repo to bob", func() {
			Expect(b.Add("https://github.com/pkg/errors.git", false)).NotTo(HaveOccurred())
		})

		It("should adds ssh repo to bob", func() {
			Expect(b.Add("git@github.com:Benchkram/bob.git", false)).NotTo(HaveOccurred())
		})

		It("should add local repo to bob", func() {
			for _, child := range childs {
				Expect(b.Add(fmt.Sprintf("file://%s", child), false)).NotTo(HaveOccurred())
			}
		})

		It("should add https repo to bob, without infering ssh url ", func() {
			targeturl := "https://github.com/pkg/requests.git"
			Expect(b.Add(targeturl, true)).NotTo(HaveOccurred())

			repo, err := getRepositoryByUrlFromBob(b, targeturl)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.SSHUrl).To(Equal(""))
			Expect(repo.LocalUrl).To(Equal(""))
			Expect(repo.HTTPSUrl).To(Equal(targeturl))
		})

		It("should add ssh repo to bob, without infering https url", func() {
			targeturl := "git@github.com:pkg/exec.git"
			Expect(b.Add(targeturl, true)).NotTo(HaveOccurred())

			repo, err := getRepositoryByUrlFromBob(b, targeturl)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.HTTPSUrl).To(Equal(""))
			Expect(repo.LocalUrl).To(Equal(""))
			Expect(repo.SSHUrl).To(Equal(targeturl))
		})

		It("should fail adding a empty url", func() {
			Expect(b.Add("", true)).To(HaveOccurred())
		})

		// It("should fail due to invalid repo name, ", func() {
		// 	err := b.Add("https://", true)
		// 	Expect(err).To(HaveOccurred())
		// })

		It("should verify that adding a duplicate repo fails", func() {
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
