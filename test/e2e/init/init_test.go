package inittest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Benchkram/bob/pkg/file"
)

var _ = Describe("Test bob init", func() {
	Context("in a fresh environment", func() {
		It("initializes bob", func() {
			Expect(b.Init()).NotTo(HaveOccurred())
		})

		It("check that .bob dir exists", func() {
			Expect(file.Exists(".bob.workspace")).To(BeTrue())
		})
	})
})
