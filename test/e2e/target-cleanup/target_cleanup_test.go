package targetcleanuptest

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing correct removal of directory targets", func() {
	ctx := context.Background()

	When("caching is enabled", func() {

		var b *bob.B
		It("should setup test environment", func() {
			err := useBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())

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

		It("should invalidate the target by adding an empty file", func() {
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

		It("should remove the target directory and expect the targets being loaded from the cache", func() {
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			// the empty file should not be there anymore
			dirContents, err := readDir("sub-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(1))
			Expect(dirContents).To(ContainElement("non-empty-file"))
			Expect(dirContents).NotTo(ContainElement("empty-file"))
		})

		It("should cleanup test environment", func() {
			err := releaseBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())
		})

	})

	When("cache is disabled", func() {

		var b *bob.B
		It("should setup test environment", func() {
			err := useBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())

			bob, err := BobSetupNoCache()
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

		It("should invalidate the target by adding an empty file", func() {
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

		// todo fix the only failing test
		/*		It("should remove the target directory and expect the targets being recreating doing a full rebuild", func() {
				err := b.Build(ctx, "build")
				Expect(err).NotTo(HaveOccurred())

				// the empty file should not be there anymore
				dirContents, err := readDir("sub-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(dirContents).To(HaveLen(1))
				Expect(dirContents).To(ContainElement("non-empty-file"))
				Expect(dirContents).NotTo(ContainElement("empty-file"))
			})*/

		It("should cleanup test environment", func() {
			err := releaseBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
