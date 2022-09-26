package targetnametest

import (
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

			Expect(err.Error()).To(Equal("duplicate target `hello` found"))
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

			Expect(err.Error()).To(Equal("duplicate target `hello` found"))
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

			Expect(err.Error()).To(Equal("duplicate target `my-image:latest` found"))
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

			Expect(err.Error()).To(Equal("duplicate target `my-image:latest` found"))
		})
	})
})
