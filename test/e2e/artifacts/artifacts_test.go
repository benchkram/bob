package artifactstest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/playbook"
	"github.com/Benchkram/bob/pkg/dockermobyutil"
	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// file targets
var _ = Describe("Test artifact and target invalidation", func() {
	Context("in a fresh playground", func() {

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
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

		// 5)
		It("should not rebuild but update the target", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 6)
		It("should not rebuild but unpack from local artifact", func() {
			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))
		})

		// 7)
		It("should rebuild and update local artifact", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())
			err = targetChange(filepath.Join(dir, "run"))
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 8)
		It("should update target from local artifact", func() {
			err := targetChange(filepath.Join(dir, "run"))
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			file.Exists(filepath.Join(dir, "run"))
		})

		It("cleanup", func() {
			err := b.CleanBuildInfoStore()
			Expect(err).NotTo(HaveOccurred())

			err = b.CleanLocalStore()
			Expect(err).NotTo(HaveOccurred())

			err = reset()
			Expect(err).NotTo(HaveOccurred())
		})

	})
})

//  docker targets
var _ = Describe("Test artifact and docker-target invalidation", func() {
	Context("in a fresh playground", func() {

		mobyClient := dockermobyutil.NewRegistryClient()

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
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

		// 5)
		It("should not rebuild but update the target", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 6)
		It("should not rebuild but unpack from local artifact", func() {
			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))
		})

		// 7)
		It("should rebuild and update local artifact", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())

			// alter the target by tagging it with a different image
			_, err = buildTask(b, bob.BuildTargetDockerImagePlusName)
			Expect(err).NotTo(HaveOccurred())
			err = mobyClient.ImageTag(
				bob.BuildTargetBobTestImagePlus,
				bob.BuildTargetBobTestImage,
			)
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 8)
		It("should update target from local artifact", func() {
			// alter the target by tagging it with a different image
			_, err := buildTask(b, bob.BuildTargetDockerImagePlusName)
			Expect(err).NotTo(HaveOccurred())
			err = mobyClient.ImageTag(
				bob.BuildTargetBobTestImagePlus,
				bob.BuildTargetBobTestImage,
			)
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("cleanup", func() {
			err := b.CleanBuildInfoStore()
			Expect(err).NotTo(HaveOccurred())

			err = b.CleanLocalStore()
			Expect(err).NotTo(HaveOccurred())

			err = reset()
			Expect(err).NotTo(HaveOccurred())

			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())
			if exists {
				err = mobyClient.ImageRemove(bob.BuildTargetBobTestImage)
				Expect(err).NotTo(HaveOccurred())
			}

			exists, err = mobyClient.ImageExists(bob.BuildTargetBobTestImagePlus)
			Expect(err).NotTo(HaveOccurred())
			if exists {
				err = mobyClient.ImageRemove(bob.BuildTargetBobTestImagePlus)
				Expect(err).NotTo(HaveOccurred())
			}
		})

	})
})
