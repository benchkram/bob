package targetsymlinktest

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing that symlink targets are preserved", func() {

	ctx := context.Background()
	When("creating & extracting artifacts", func() {

		var b *bob.B
		It("should setup test environment", func() {
			err := useBobfile("with_symlink")
			Expect(err).NotTo(HaveOccurred())

			bob, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			b = bob
		})

		It("should build the task", func() {
			err := b.Build(ctx, "build")
			errz.Log(err)
			Expect(err).NotTo(HaveOccurred())

			dirContents, err := readDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(HaveLen(3))
			Expect(dirContents).To(ContainElement("hello"))
			Expect(dirContents).To(ContainElement("shortcut"))
			Expect(dirContents).To(ContainElement("bob.yaml"))
		})

		It("should invalidate the target to trigger a load from the artifact", func() {
			err := os.Remove("shortcut")
			Expect(err).NotTo(HaveOccurred())
			err = os.Remove("hello")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should load the targets from the cache and verify the symlink integrity", func() {
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			// the contents of dir should stay the same
			dirContents, err := readDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(dirContents).To(ContainElement("hello"))
			Expect(dirContents).To(ContainElement("shortcut"))

			isSimlink, err := file.IsSymlink("./shortcut")
			Expect(err).NotTo(HaveOccurred())
			Expect(isSimlink).To(BeTrue())

			src, err := os.Readlink("./shortcut")
			Expect(err).NotTo(HaveOccurred())
			Expect(src).To(Equal("hello"))
		})

		It("should cleanup test environment", func() {
			err := releaseBobfile("with_symlink")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
