package artifactstest

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// file targets
var _ = Describe("Test artifact and target lifecycle without existing buildinfo", func() {
	Context("in a fresh playground", func() {

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("should build", func() {
			ctx := context.Background()
			err := b.Build(ctx, "build")
			errz.Log(err)
			Expect(err).NotTo(HaveOccurred())
		})

		var artifactID string
		It("should check that exactly one artifact was created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))
			artifactID = fs[0].Name()
		})

		It("clean artifacts & buildinfo", func() {
			err := b.CleanLocalStore()
			Expect(err).NotTo(HaveOccurred())
		})

		// 1)
		It("should rebuild, update the target and write the artifact", func() {
			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("clean artifacts & buildinfo", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
		})

		// 2) 4)
		It("should not rebuild as a artifact exists", func() {
			// create artifact
			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))
			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// clean buildinfo
			err = b.CleanBuildInfoStore()
			Expect(err).NotTo(HaveOccurred())

			state, err = buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))
		})

		It("clean artifacts & buildinfo", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
		})

		// 3)
		It("should rebuild as no artifact exists", func() {
			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("cleanup", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
			err = reset()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

// docker targets
var _ = Describe("Test artifact and target lifecycle for docker images without existing buildinfo", func() {
	Context("in a fresh playground", func() {

		mobyClient := dockermobyutil.NewRegistryClient()

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("should build", func() {
			ctx := context.Background()
			err := b.Build(ctx, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
		})

		var artifactID string
		It("should check that exactly one artifact was created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))
			artifactID = fs[0].Name()
		})

		It("clean artifacts & buildinfo", func() {
			err := b.CleanLocalStore()
			Expect(err).NotTo(HaveOccurred())
		})

		// 1)
		It("should rebuild, update the target and write the artifact", func() {
			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			exists, err = mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("clean artifacts & buildinfo", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
		})

		// 2) 4)
		It("should not rebuild as a artifact exists", func() {
			// create artifact
			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))
			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// clean buildinfo
			err = b.CleanBuildInfoStore()
			Expect(err).NotTo(HaveOccurred())

			state, err = buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))
		})

		It("clean artifacts & buildinfo", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
		})

		// 3)
		It("should rebuild as no artifact exists", func() {
			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("cleanup", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())
			err = reset()
			Expect(err).NotTo(HaveOccurred())

			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())

			if exists {
				err = mobyClient.ImageRemove(bob.BuildTargetBobTestImage)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
