package buildtest

import (
	"context"
	"errors"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/file"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob build", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("runs a slow build and cancelles the context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := b.Build(ctx, "slow")
			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, context.Canceled)).To(BeTrue())

			Expect(file.Exists("slowdone")).To(BeFalse(), "slowdone file shouldn't exist")
		})

		It("runs a slow build without cancelling it", func() {
			ctx := context.Background()
			Expect(b.Build(ctx, "slow")).NotTo(HaveOccurred())

			Expect(file.Exists("slowdone")).To(BeTrue(), "slowdone file should exist")
		})
	})
})
