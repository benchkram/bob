package ignoretest

import (
	"context"
	"io/fs"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/playbook"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob build", func() {
	Context("in a fresh environment", func() {
		const (
			taskname   = "ignoredInputs"
			inputfile  = "fileToWatch"
			ignorefile = "fileToIgnore"
		)
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("create files for build task", func() {
			someText := []byte("hello this is a text in a file\n")

			err := os.WriteFile(inputfile, someText, fs.ModePerm)
			Expect(err).NotTo(HaveOccurred())
			err = os.WriteFile(ignorefile, someText, fs.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("needs to be built the first time", func() {
			ctx := context.Background()
			Expect(b.Build(ctx, taskname)).NotTo(HaveOccurred())
		})

		It("needs no rebuild after that", func() {
			checkTaskState(b, taskname, playbook.StateNoRebuildRequired)
		})

		It("change of ignored file doesn't need rebuild", func() {
			someText := []byte("another line\n")

			err := os.WriteFile(ignorefile, someText, fs.ModeAppend)
			Expect(err).NotTo(HaveOccurred())

			checkTaskState(b, taskname, playbook.StateNoRebuildRequired)
		})

		It("change of input file requires rebuild", func() {
			someText := []byte("another line\n")

			err := os.WriteFile(inputfile, someText, fs.ModeAppend)
			Expect(err).NotTo(HaveOccurred())

			// Check for completed since this function will call rerun
			// Therefor this task will be in state completed afterwards
			checkTaskState(b, taskname, playbook.StateCompleted)
		})
	})
})

func checkTaskState(b *bob.B, taskname string, expectedState playbook.State) {
	aggregate, err := b.Aggregate()
	Expect(err).NotTo(HaveOccurred())
	pb, err := aggregate.Playbook(taskname)
	Expect(err).NotTo(HaveOccurred())

	err = pb.Build(context.Background())
	Expect(err).NotTo(HaveOccurred())

	ts, err := pb.TaskStatus(taskname)
	Expect(err).NotTo(HaveOccurred())

	Expect(ts.State()).To(Equal(expectedState))
}
