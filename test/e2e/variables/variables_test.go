package variablestest

import (
	"context"

	"github.com/Benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob variable usage", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("runs a task which uses a variable", func() {
			Expect(b.Build(context.Background(), "print")).NotTo(HaveOccurred())
		})
	})
})
