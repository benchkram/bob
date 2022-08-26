package inputest

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing input for a task", func() {
	BeforeEach(func() {
		testDir, err := ioutil.TempDir("", "input-test-*")
		Expect(err).NotTo(HaveOccurred())

		dir = testDir
		err = os.Chdir(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("input is * and task depends on another task with a target", func() {
		It("should not include the children tasks targets in input", func() {
			func() {
				bf, ok := nameToBobfile["with_one_level"]
				Expect(ok).To(BeTrue())

				err := bf.BobfileSave(dir, "bob.yaml")
				Expect(err).NotTo(HaveOccurred())
			}()

			defer func() {
				err := os.Remove("bob.yaml")
				Expect(err).NotTo(HaveOccurred())
			}()

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			taskName := "build"

			bobfile, err := b.Aggregate()
			errz.Log(err)
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

	When("input is * and task depends on another tasks with a target from a second-level dir", func() {
		It("should not include the children tasks targets in input", func() {
			func() {
				bf, ok := nameToBobfile["with_second_level"]
				Expect(ok).To(BeTrue())
				err := bf.BobfileSave(dir, "bob.yaml")
				Expect(err).NotTo(HaveOccurred())

				err = os.Mkdir(filepath.Join(dir, "second_level"), 0700)
				Expect(err).NotTo(HaveOccurred())

				secondBf, ok := nameToBobfile["with_second_level/second_level"]
				Expect(ok).To(BeTrue())
				err = secondBf.BobfileSave(filepath.Join(dir, "second_level"), "bob.yaml")
				Expect(err).NotTo(HaveOccurred())
			}()

			defer func() {
				err := os.Remove("bob.yaml")
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(filepath.Join(dir, "second_level", "bob.yaml"))
				Expect(err).NotTo(HaveOccurred())
			}()

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			taskName := "build"

			bobfile, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok := bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsBeforeBuild := task.Inputs()
			Expect(len(inputsBeforeBuild)).To(Equal(2))
			Expect(inputsBeforeBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))
			Expect(inputsBeforeBuild[1]).To(Equal(filepath.Join(dir, "second_level", "bob.yaml")))

			// Build and aggregate again
			err = b.Build(ctx, taskName)
			Expect(err).NotTo(HaveOccurred())
			bobfile, err = b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok = bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsAfterBuild := task.Inputs()
			Expect(len(inputsAfterBuild)).To(Equal(2))
			Expect(inputsAfterBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))
			Expect(inputsAfterBuild[1]).To(Equal(filepath.Join(dir, "second_level", "bob.yaml")))
		})
	})

	When("input is * and task depends on another tasks with a target from a third-level dir", func() {
		It("should not include the children tasks targets in input", func() {
			func() {
				bf, ok := nameToBobfile["with_three_level"]
				Expect(ok).To(BeTrue())
				err := bf.BobfileSave(dir, "bob.yaml")
				Expect(err).NotTo(HaveOccurred())

				err = os.Mkdir(filepath.Join(dir, "second_level"), 0700)
				Expect(err).NotTo(HaveOccurred())

				secondBf, ok := nameToBobfile["with_three_level/second_level"]
				Expect(ok).To(BeTrue())
				err = secondBf.BobfileSave(filepath.Join(dir, "second_level"), "bob.yaml")
				Expect(err).NotTo(HaveOccurred())

				err = os.Mkdir(filepath.Join(dir, "second_level", "third_level"), 0700)
				Expect(err).NotTo(HaveOccurred())

				thirdLevelBf, ok := nameToBobfile["with_three_level/second_level/third_level"]
				Expect(ok).To(BeTrue())
				err = thirdLevelBf.BobfileSave(filepath.Join(dir, "second_level", "third_level"), "bob.yaml")
				Expect(err).NotTo(HaveOccurred())
			}()

			defer func() {
				err := os.Remove("bob.yaml")
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(filepath.Join(dir, "second_level", "bob.yaml"))
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(filepath.Join(dir, "second_level", "third_level", "bob.yaml"))
				Expect(err).NotTo(HaveOccurred())
			}()

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())
			taskName := "build"

			bobfile, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok := bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsBeforeBuild := task.Inputs()
			Expect(len(inputsBeforeBuild)).To(Equal(3))
			Expect(inputsBeforeBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))
			Expect(inputsBeforeBuild[1]).To(Equal(filepath.Join(dir, "second_level", "bob.yaml")))
			Expect(inputsBeforeBuild[2]).To(Equal(filepath.Join(dir, "second_level", "third_level", "bob.yaml")))

			// Build and aggregate again
			err = b.Build(ctx, taskName)
			Expect(err).NotTo(HaveOccurred())
			bobfile, err = b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			task, ok = bobfile.BTasks[taskName]
			Expect(ok).To(BeTrue())

			inputsAfterBuild := task.Inputs()
			Expect(len(inputsAfterBuild)).To(Equal(3))
			Expect(inputsAfterBuild[0]).To(Equal(filepath.Join(dir, "bob.yaml")))
			Expect(inputsAfterBuild[1]).To(Equal(filepath.Join(dir, "second_level", "bob.yaml")))
			Expect(inputsAfterBuild[2]).To(Equal(filepath.Join(dir, "second_level", "third_level", "bob.yaml")))
		})
	})
})
