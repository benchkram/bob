package targetcleanuptest

import (
	"os"

	"github.com/benchkram/bob/bob"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing correct removal of directory targets", func() {

	When("caching is enabled", func() {

		var b *bob.B
		It("should setup test environment", func() {
			useBobfile("with_dir_target")
			defer releaseBobfile("with_dir_target")

			bob, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			b = bob
		})

		It("should build the task", func() {
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			dirContents, err := readDir("sub-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(1))
			Expect(dirContents).To(ContainElement("non-empty-file"))
		})

		It("should invalidate the target", func() {
			// create an empty file inside sub-dir
			emptyFile, err := os.Create("./sub-dir/empty-file")
			Expect(err).NotTo(HaveOccurred())
			err = emptyFile.Close()
			Expect(err).NotTo(HaveOccurred())

			dirContents, err := readDir("sub-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(2))
			Expect(dirContents).To(ContainElement("non-empty-file"))
			Expect(dirContents).To(ContainElement("empty-file"))
		})

		It("should rebuild the task and expect the targets beeing loaded from the cache", func() {
			// re-build
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			// the empty file should not be there anymore
			dirContents, err := readDir("sub-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(1))
			Expect(dirContents).To(ContainElement("non-empty-file"))
			Expect(dirContents).NotTo(ContainElement("empty-file"))
		})

	})

	// TODO: GOON

	Context("cache is disabled", func() {
		When("cache is disabled", func() {
			It("removes and reload from the cache the dir target", func() {
				useBobfile("with_dir_target")
				defer releaseBobfile("with_dir_target")

				b, err := BobSetupNoCache()
				Expect(err).NotTo(HaveOccurred())

				err = b.Build(ctx, "build")
				Expect(err).NotTo(HaveOccurred())

				dirContents, err := readDir("sub-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dirContents).To(HaveLen(1))
				Expect(dirContents).To(ContainElement("non-empty-file"))

				// create an empty file inside sub-dir
				emptyFile, err := os.Create("./sub-dir/empty-file")
				Expect(err).NotTo(HaveOccurred())
				err = emptyFile.Close()
				Expect(err).NotTo(HaveOccurred())

				dirContents, err = readDir("sub-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dirContents).To(HaveLen(2))
				Expect(dirContents).To(ContainElement("non-empty-file"))
				Expect(dirContents).To(ContainElement("empty-file"))

				// re-build
				err = b.Build(ctx, "build")
				Expect(err).NotTo(HaveOccurred())

				// the empty file should not be there anymore
				dirContents, err = readDir("sub-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dirContents).To(HaveLen(1))
				Expect(dirContents).To(ContainElement("non-empty-file"))
				Expect(dirContents).NotTo(ContainElement("empty-file"))
			})
		})
	})
})
