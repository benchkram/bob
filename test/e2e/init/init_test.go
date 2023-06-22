package inittest

import (
	"errors"

	"github.com/benchkram/bob/bob"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/benchkram/bob/pkg/file"
)

var _ = Describe("Test bob init", func() {
	Context("in a fresh environment", func() {
		It("initializes bob", func() {
			Expect(b.Init()).NotTo(HaveOccurred())
		})

		It("check that .bob dir exists", func() {
			Expect(file.Exists(".bob.workspace")).To(BeTrue())
		})

		It("check that bob fails gracefully if .bob already exists", func() {
			err := b.Init()
			Expect(err).To(HaveOccurred())
			unwrappedErr := errors.Unwrap(err)
			Expect(unwrappedErr).NotTo(BeNil())
			Expect(errors.Is(unwrappedErr, bob.ErrWorkspaceAlreadyInitialised)).To(BeTrue())
		})
	})
})
