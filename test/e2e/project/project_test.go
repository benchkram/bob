package projectest

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing project name for tats", func() {
	Context("when project name is not set", func() {
		It("all tasks should contain working dir", func() {
			useBobFile("without_project_name")
			defer releaseBobfile("without_project_name")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			ag, err := b.AggregateSparse()
			Expect(err).NotTo(HaveOccurred())

			for _, v := range ag.BTasks {
				Expect(v.Project()).To(Equal(dir))
			}
		})
	})

	Context("when project name is set", func() {
		It("all tasks should have project the name set", func() {
			useBobFile("with_project_name")
			defer releaseBobfile("with_project_name")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			ag, err := b.AggregateSparse()
			Expect(err).NotTo(HaveOccurred())

			for _, v := range ag.BTasks {
				Expect(v.Project()).To(Equal("projectX"))
			}
		})
	})

	Context("when project name is set and there is a second level bob file", func() {
		It("all tasks should have project the name set", func() {
			Expect(os.Rename("with_second_level.yaml", "bob.yaml")).NotTo(HaveOccurred())
			Expect(os.Rename("with_second_level_second_level.yaml", dir+"/second_level/bob.yaml")).NotTo(HaveOccurred())

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			ag, err := b.AggregateSparse()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(ag.BTasks)).To(Equal(2))

			for _, v := range ag.BTasks {
				Expect(v.Project()).To(Equal("projectWithSecondLevel"))
			}
		})
	})
})
