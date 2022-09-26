package taskdecorationtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing decoration of imported tasks", func() {
	When("a task from second level gets decorated", func() {
		It("will update that task with correct dependsOn", func() {
			useBobfile("with_decoration")
			defer releaseBobfile("with_decoration")

			useSecondLevelBobfile("with_decoration")
			defer releaseSecondLevelBobfile("with_decoration")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			ag, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(ag.BTasks)).To(Equal(3))

			decoratedTask, ok := ag.BTasks["second/build"]
			Expect(ok).To(BeTrue())
			Expect(decoratedTask.DependsOn).To(Equal([]string{"before", "second/hello"}))
		})
	})

	When("a task from second and third level get decorated", func() {
		It("will update decorated task with correct dependsOn", func() {
			useBobfile("with_thirdlevel_decoration")
			defer releaseBobfile("with_thirdlevel_decoration")

			useSecondLevelBobfile("with_thirdlevel_decoration")
			defer releaseSecondLevelBobfile("with_thirdlevel_decoration")

			useThirdLevelBobfile("with_thirdlevel_decoration")
			defer releaseThirdLevelBobfile("with_thirdlevel_decoration")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			ag, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(ag.BTasks)).To(Equal(6))

			decoratedTaskSecond, ok := ag.BTasks["second/build"]
			Expect(ok).To(BeTrue())
			Expect(decoratedTaskSecond.DependsOn).To(Equal([]string{"before", "second/create"}))

			decoratedTaskThird, ok := ag.BTasks["second/third/build"]
			Expect(ok).To(BeTrue())
			Expect(decoratedTaskThird.DependsOn).To(Equal([]string{"second/hello", "second/third/hello"}))
		})
	})

	When("attempting to decorate a missing child task", func() {
		It("should fail with task does not exist error", func() {
			useBobfile("with_missed_decoration")
			defer releaseBobfile("with_missed_decoration")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("task `second/build` does not exist"))
		})
	})

	When("a decorated task contain more than dependsOn node", func() {
		It("should fail because with an error", func() {
			useBobfile("with_invalid_decoration")
			defer releaseBobfile("with_invalid_decoration")

			useSecondLevelBobfile("with_invalid_decoration")
			defer releaseSecondLevelBobfile("with_invalid_decoration")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("task `second/build` modifies an imported task. It can only contain a `dependsOn` property"))
		})
	})
})
