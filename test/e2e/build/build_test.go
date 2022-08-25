package buildtest

import (
	"context"
	"errors"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/pkg/file"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob build", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("runs a slow build and cancelles the context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := b.Build(ctx, "slow")
			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, context.Canceled)).To(BeTrue())

			Expect(file.Exists("slowdone")).To(BeFalse(), "slowdone file shouldn't exist")
		})

		It("runs a slow build without cancelling it", func() {
			ctx := context.Background()
			Expect(b.Build(ctx, "slow")).NotTo(HaveOccurred())

			Expect(file.Exists("slowdone")).To(BeTrue(), "slowdone file should exist")
		})

		It("expect rebuild always true without change for always rebuild task", func() {

			ctx := context.Background()

			targetTask := bob.BuildAlwaysTargetName
			Expect(b.Build(ctx, targetTask)).NotTo(HaveOccurred())

			aggregate, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())
			pb, err := aggregate.Playbook(targetTask)
			Expect(err).NotTo(HaveOccurred())

			task := pb.Tasks[targetTask]
			hashIn, err := task.HashIn()
			Expect(err).NotTo(HaveOccurred())

			rebuildRequired, rebuildCause, err := pb.TaskNeedsRebuild(task.Name(), hashIn)
			Expect(err).NotTo(HaveOccurred())

			Expect(rebuildRequired).To(BeTrue())
			Expect(rebuildCause).To(Equal(playbook.TaskForcedRebuild))
		})
	})
})
