package inputest

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing input for a task on one-level", func() {
	When("input is * and it depends on some tasks with a target", func() {
		It("should not include the children tasks targets in input", func() {
			useBobfile("with_one_level")
			defer releaseBobfile("with_one_level")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			taskName := "build"

			bobfile, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok := bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsBeforeBuild := task.Inputs()
			Expect(len(inputsBeforeBuild)).To(Equal(1))
			Expect(inputsBeforeBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))

			// Build and aggregate again
			err = b.Build(ctx, taskName)
			Expect(err).NotTo(HaveOccurred())
			bobfile, err = b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok = bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsAfterBuild := task.Inputs()
			Expect(len(inputsAfterBuild)).To(Equal(1))
			Expect(inputsAfterBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))
		})
	})
})
