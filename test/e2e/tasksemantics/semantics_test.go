package tasksemanticstest

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob/playbook"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing task semantics behaviour", func() {
	When("task has input * and a target", func() {
		b, err := Bob()
		Expect(err).NotTo(HaveOccurred())

		It("it will first build successfully", func() {
			useBobfile("rebuild_on_input_change")
			defer releaseBobfile("rebuild_on_input_change")

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})

		It("it will have summary cached on the second build", func() {
			capture()
			useBobfile("rebuild_on_input_change")
			defer releaseBobfile("rebuild_on_input_change")

			ctx := context.Background()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCached := playbook.StateCached
			Expect(output()).To(ContainSubstring(buildCached.Summary()))
		})

		It("it will be rebuilt when input * changes", func() {
			useBobfile("rebuild_on_input_change")
			defer releaseBobfile("rebuild_on_input_change")

			// invalidate target by creating a new file
			err := os.WriteFile("./someFile", []byte("hello\ngo\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})
	})

	When("task has target but no input set", func() {
		It("it will trigger a no input provided error on aggregation", func() {
			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			useBobfile("no_input_with_target")
			defer releaseBobfile("no_input_with_target")

			_, err = b.Aggregate()
			Expect(err.Error()).To(Equal("no input provided for task `build`"))
		})
	})

	// An unknown target is a file created inside `cmd` which is not specified in `target`
	When("task has no input and an unknown target to be created", func() {
		It("it will trigger a no input provided error on aggregation", func() {
			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			useBobfile("no_input_unknown_target")
			defer releaseBobfile("no_input_unknown_target")

			_, err = b.Aggregate()
			Expect(err.Error()).To(Equal("no input provided for task `build`"))
		})
	})

	When("task has no input and no target defined", func() {
		It("it will trigger a no input provided error on aggregation", func() {
			useBobfile("no_input_no_target")
			defer releaseBobfile("no_input_no_target")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err.Error()).To(Equal("no input provided for task `build`"))
		})

		It("it will not trigger a no input provided error if rebuild:always", func() {
			useBobfile("no_input_no_target_rebuild_always")
			defer releaseBobfile("no_input_no_target_rebuild_always")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			_, err = b.Aggregate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("task has input * and no target to be created", func() {
		b, err := Bob()
		Expect(err).NotTo(HaveOccurred())

		It("it will first build successfully", func() {
			useBobfile("with_input_no_target")
			defer releaseBobfile("with_input_no_target")

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})

		It("it will have summary no-rebuild on second build", func() {
			useBobfile("with_input_no_target")
			defer releaseBobfile("with_input_no_target")

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCached := playbook.StateNoRebuildRequired
			Expect(output()).To(ContainSubstring(buildCached.Summary()))
		})

		It("it will be rebuilt when input * changes", func() {
			useBobfile("with_input_no_target")
			defer releaseBobfile("with_input_no_target")

			// invalidate target by creating a new file
			err := os.WriteFile("./someRandomFile", []byte("hello\ngo\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})
	})

	When("task is a compound task", func() {
		b, err := Bob()
		Expect(err).NotTo(HaveOccurred())

		It("it will first build successfully", func() {
			useBobfile("compound_task")
			defer releaseBobfile("compound_task")

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})

		It("it will have summary no-rebuild on second build", func() {
			useBobfile("compound_task")
			defer releaseBobfile("compound_task")

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			// generated task is cached
			buildCached := playbook.StateCached
			out := output()
			Expect(out).To(ContainSubstring(buildCached.Summary()))

			// build is marked as no-rebuild
			buildNotRequired := playbook.StateNoRebuildRequired
			Expect(out).To(ContainSubstring(buildNotRequired.Summary()))
		})

		It("it will be rebuilt when input * changes", func() {
			useBobfile("compound_task")
			defer releaseBobfile("compound_task")

			// invalidate target by creating a new file
			err := os.WriteFile("./anotherRandomFile", []byte("hello\ngo\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			buildCompleted := playbook.StateCompleted
			Expect(output()).To(ContainSubstring(buildCompleted.Summary()))
		})
	})

	When("task has rebuild:always and target set", func() {
		It("it will trigger an error on aggregation", func() {
			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			useBobfile("rebuild_always_with_target")
			defer releaseBobfile("rebuild_always_with_target")

			_, err = b.Aggregate()
			Expect(err.Error()).To(Equal("`rebuild:always` not allowed in combination with `target` for task: `build`"))
		})
	})
})
