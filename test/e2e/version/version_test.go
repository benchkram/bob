package version_test

import (
	"context"
	"github.com/Benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob's file export validation", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground with bob v1.0.0", func() {
			bob.Version = "1.0.0"

			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("run verify and make sure warnings are shown", func() {
			capture()

			err := b.Verify(context.Background())
			Expect(err).NotTo(HaveOccurred())

			out := output()
			Expect(out).To(ContainSubstring("major version mismatch"))
			Expect(out).To(ContainSubstring("possible version incompatibility"))
		})
	})
})
