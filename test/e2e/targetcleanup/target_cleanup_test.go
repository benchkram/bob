package targetcleanuptest

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var cacheEnabled bool

var _ = Describe("Testing cleaning up dir targets", func() {
	Context("cache is enabled", func() {
		cacheEnabled = true
		When("a rebuild is done", func() {
			It("removes and reload from the cache the dir target", func() {
				useBobfile("with_dir_target")
				defer releaseBobfile("with_dir_target")
				b, err := BobSetup(cacheEnabled)
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

	Context("cache is disabled", func() {
		cacheEnabled = false
		When("a rebuild is done", func() {
			It("removes and reload from the cache the dir target", func() {
				useBobfile("with_dir_target")
				defer releaseBobfile("with_dir_target")

				b, err := BobSetup(cacheEnabled)
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
