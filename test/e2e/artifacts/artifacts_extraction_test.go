package artifactstest

import (
	"context"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/bob/pkg/dockermobyutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test artifact creation and extraction", func() {
	Context("in a fresh playground", func() {

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("assure to be in the test root dir", func() {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(wd).To(Equal(dir))
		})

		It("should build", func() {
			ctx := context.Background()
			err := b.Build(ctx, bob.BuildTargetwithdirsTargetName)
			Expect(err).NotTo(HaveOccurred())
		})

		var artifactID string
		It("should check that exactly one artifact was created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))
			artifactID = fs[0].Name()
			println(artifactID)
		})

		It("inspect artifact", func() {
			artifact, err := artifactStore.GetArtifact(context.Background(), artifactID)
			Expect(err).NotTo(HaveOccurred())
			description, err := bobtask.ArtifactInspectFromReader(artifact)
			Expect(err).NotTo(HaveOccurred())
			Expect(description.Metadata().TargetType).To(Equal(target.Path))

			println(description)
		})

		It("cleanup build/target dir", func() {
			err := os.RemoveAll(".bbuild")
			Expect(err).NotTo(HaveOccurred())
		})

		It("extract artifact from store on rebuild", func() {
			err := b.Build(context.Background(), bob.BuildTargetwithdirsTargetName)
			Expect(err).NotTo(HaveOccurred())
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

var _ = Describe("Test artifact creation and extraction from docker targets", func() {
	Context("in a fresh playground", func() {

		mobyClient := dockermobyutil.NewRegistryClient()

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("should assure test image is not in docker registry", func() {
			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())

			if exists {
				err = mobyClient.ImageRemove(bob.BuildTargetBobTestImage)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("assure to be in the test root dir", func() {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(wd).To(Equal(dir))
		})

		It("should build", func() {
			ctx := context.Background()
			err := b.Build(ctx, bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should check that the docker image was created correctly", func() {
			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		var artifactID string
		It("should check that exactly one artifact was created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))
			artifactID = fs[0].Name()
			println(artifactID)
		})

		It("inspect artifact", func() {
			artifact, err := artifactStore.GetArtifact(context.Background(), artifactID)
			Expect(err).NotTo(HaveOccurred())
			description, err := bobtask.ArtifactInspectFromReader(artifact)
			Expect(err).NotTo(HaveOccurred())
			Expect(description.Metadata().TargetType).To(Equal(target.Docker))
		})

		It("should remove test image from docker registry", func() {
			exists, err := mobyClient.ImageExists(bob.BuildTargetBobTestImage)
			Expect(err).NotTo(HaveOccurred())

			if exists {
				err = mobyClient.ImageRemove(bob.BuildTargetBobTestImage)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should extract artifact from store on rebuild", func() {
			err := b.Build(context.Background(), bob.BuildTargetDockerImageName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should check that the docker image was created correctly", func() {
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
		})

	})
})
