package artifactstest

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test artifact and target on no cache build", func() {
	Context("in a fresh playground", func() {

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("should build both with and without cache", func() {
			ctx := context.Background()

			err := bNoCache.Build(ctx, "build")
			errz.Log(err)

			Expect(err).NotTo(HaveOccurred())
		})

		It("no artifacts should be created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(0))
		})

		It("should rebuild and create a new artifact for always-build without no-cache", func() {
			// b and bNoCache uses the same dir, so this should
			// work without running b.Build first
			state, err := buildTask(b, bob.BuildAlwaysTargetName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))

			artifactID := fs[0].Name()
			err = artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should rebuild and create no artifact for always-build with no-cache", func() {
			ctx := context.Background()
			err := bNoCache.Build(ctx, bob.BuildAlwaysTargetName)
			Expect(err).NotTo(HaveOccurred())

			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(0))
		})

		It("should rebuild and create no artifact for all task build with no-cache", func() {
			ctx := context.Background()
			err := bNoCache.Build(ctx, bob.BuildAllTargetName)
			Expect(err).NotTo(HaveOccurred())

			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(0))
		})

		It("cleanup", func() {
			// b and bNoCache uses the same dir, so this should
			// so this cleans directory for both build task
			err := bNoCache.CleanBuildInfoStore()
			Expect(err).NotTo(HaveOccurred())

			err = bNoCache.CleanLocalStore()
			Expect(err).NotTo(HaveOccurred())

			err = reset()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
