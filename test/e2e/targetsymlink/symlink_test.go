package targetsymlinktest

import (
	"os"

	"github.com/benchkram/bob/pkg/file"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing targets with symlink", func() {
	When("a rebuild is done", func() {
		It("the target file should preserve its symlink", func() {
			useBobfile("with_symlink")
			defer releaseBobfile("with_symlink")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			dirContents, err := contentsOfDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(3))
			Expect(dirContents).To(ContainElement("hello"))
			Expect(dirContents).To(ContainElement("shortcut"))
			Expect(dirContents).To(ContainElement("bob.yaml"))

			// re-build
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			// the contents of dir should stay the same
			dirContents, err = contentsOfDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(3))
			Expect(dirContents).To(ContainElement("hello"))
			Expect(dirContents).To(ContainElement("shortcut"))
			Expect(dirContents).To(ContainElement("bob.yaml"))

			isSimlink, err := file.IsSymlink("./shortcut")
			Expect(err).NotTo(HaveOccurred())
			Expect(isSimlink).To(BeTrue())

			src, err := os.Readlink("./shortcut")
			Expect(err).NotTo(HaveOccurred())
			Expect(src).To(Equal("hello"))
		})
	})
})
