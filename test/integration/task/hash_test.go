package tasktest

import (
	"github.com/Benchkram/bob/bob/build"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test task-related functionality", func() {
	Context("in a fresh environment", func() {
		const (
			shouldInfluenceText = "dummy text which should influence the exported structure"
		)

		It("should produce the same hash for two tasks with the same exported structure", func() {
			t1 := build.Make()
			t1h, err := t1.Hash()
			Expect(err).NotTo(HaveOccurred())

			t2 := build.Make()
			t2h, err := t2.Hash()
			Expect(err).NotTo(HaveOccurred())

			Expect(t1h).To(Equal(t2h))
		})

		It("should produce different hashes for two tasks with a different exported structure", func() {
			t1 := build.Make()
			t1h, err := t1.Hash()
			Expect(err).NotTo(HaveOccurred())

			t2 := build.Make()
			t2.InputDirty.Inputs = []string{shouldInfluenceText}
			t2h, err := t2.Hash()
			Expect(err).NotTo(HaveOccurred())

			Expect(t1h).NotTo(Equal(t2h))
		})

		It("should produce the same hash for two taskst wih same environment", func() {
			t1 := build.Make(build.WithEnvironment([]string{
				"PATH=/usr/bin",
			}))
			t1h, err := t1.Hash()
			Expect(err).NotTo(HaveOccurred())

			t2 := build.Make(build.WithEnvironment([]string{
				"PATH=/usr/bin",
			}))
			t2h, err := t2.Hash()
			Expect(err).NotTo(HaveOccurred())
			Expect(t1h).To(Equal(t2h))

		})

		It("should produce different hashes for two tasks with a different environments", func() {
			t1 := build.Make(build.WithEnvironment([]string{
				"PATH=/usr/bin",
			}))
			t1h, err := t1.Hash()
			Expect(err).NotTo(HaveOccurred())

			t2 := build.Make(build.WithEnvironment([]string{
				"PATH=/usr/etc",
			}))
			t2h, err := t2.Hash()
			Expect(err).NotTo(HaveOccurred())

			Expect(t1h).NotTo(Equal(t2h))
		})

	})
})
