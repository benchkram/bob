package targetnametest

import (
	"github.com/benchkram/bob/bobtask"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing name of targets", func() {
	When("two tasks from same bobfile have the same target", func() {
		It("will return an error on aggregation", func() {
			useBobfile("with_simple_file")
			defer releaseBobfile("with_simple_file")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(HaveOccurred())

			expected := bobtask.CreateErrAmbiguousTargets([]string{"another", "build"}, "hello").Error()
			Expect(err.Error()).To(Equal(expected))
		})
	})

	When("two tasks(one containing multiline targets) from same bobfile have the same target", func() {
		It("will return an error on aggregation", func() {
			useBobfile("with_multiline_target")
			defer releaseBobfile("with_multiline_target")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(HaveOccurred())

			expected := bobtask.CreateErrAmbiguousTargets([]string{"another", "build"}, "hello").Error()
			Expect(err.Error()).To(Equal(expected))
		})
	})

	When("a target with same name exists in a second level bob file", func() {
		It("should not return error on aggregation because they have different paths", func() {
			useBobfile("with_second_level")
			defer releaseBobfile("with_second_level")

			useSecondLevelBobfile("with_second_level")
			defer releaseSecondLevelBobfile("with_second_level")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	When("two tasks from same bobfile have the same image target", func() {
		It("will return an error on aggregation", func() {
			useBobfile("with_target_image")
			defer releaseBobfile("with_target_image")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(HaveOccurred())

			expected := bobtask.CreateErrAmbiguousTargets([]string{"another", "build"}, "my-image:latest").Error()
			Expect(err.Error()).To(Equal(expected))
		})
	})

	When("an image target with same name exists in a second level bob file", func() {
		It("should return error on aggregation", func() {
			useBobfile("with_target_image_second_level")
			defer releaseBobfile("with_target_image_second_level")

			useSecondLevelBobfile("with_target_image_second_level")
			defer releaseSecondLevelBobfile("with_target_image_second_level")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(HaveOccurred())

			expected := bobtask.CreateErrAmbiguousTargets([]string{"build", "second/build"}, "my-image:latest").Error()
			Expect(err.Error()).To(Equal(expected))
		})
	})

	When("a target with same path exists in a second level bob file", func() {
		It("should return error on aggregation because targets are identical", func() {
			useBobfile("with_second_level_same_path")
			defer releaseBobfile("with_second_level_same_path")

			useSecondLevelBobfile("with_second_level_same_path")
			defer releaseSecondLevelBobfile("with_second_level_same_path")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).To(HaveOccurred())

			expected := bobtask.CreateErrAmbiguousTargets([]string{"build", "second/build"}, "second/hello").Error()
			Expect(err.Error()).To(Equal(expected))
		})
	})
})
