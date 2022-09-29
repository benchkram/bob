package targetcleanuptest

import (
	"context"

	"github.com/benchkram/bob/bob"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing reuse of targets in other tasks", func() {
	ctx := context.Background()

	When("cache is enabled", func() {

		var b *bob.B
		It("should setup test environment", func() {
			err := useBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())

			bob, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			b = bob
		})

		It("should build the task", func() {
			err := b.Build(ctx, "cat")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should build the task again but reuse the artifact from the cache", func() {
			err := b.Build(ctx, "cat")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should cleanup test environment", func() {
			err := releaseBobfile("with_dir_target")
			Expect(err).NotTo(HaveOccurred())
		})

	})

})
