package nix_test

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/filepathutil"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing new nix implementation", func() {
	AfterEach(func() {
		filepathutil.ClearListRecursiveCache()
	})
	Context("with use-nix false", func() {
		It("build without errors", func() {
			bob.Version = "1.0.0"
			// update bob.yaml with mock content
			err := os.Rename("with_use_nix_false.yaml", "bob.yaml")
			errz.Log(err)

			ctx := context.Background()
			err = b.Build(ctx, "build")
			errz.Log(err)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with task dependencies", func() {
		It("task dependency go_1_17 will have priority to bob file go_1_18", func() {
			Expect(os.Rename("with_task_dependencies.yaml", "bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.17.9"))
		})
	})

	Context("with dependencies per bob file", func() {
		It("running task go version will use go_1_16 from bob file dependency", func() {
			Expect(os.Rename("with_bob_dependencies.yaml", "bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.16.15"))
		})
	})

	Context("with ambiguous deps in root", func() {
		It("running task go version will use go_1_17 from bob file dependency", func() {
			Expect(os.Rename("with_ambiguous_deps_in_root.yaml", "bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.17"))
		})
	})

	Context("with ambiguous deps in task", func() {
		It("running task go version will use go_1_17 from task deps", func() {
			Expect(os.Rename("with_ambiguous_deps_in_task.yaml", "bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.17"))
		})
	})

	Context("with second level bob file", func() {
		It("running task go version from parent directory will use go_1_17 second level task dependencies", func() {
			Expect(os.Rename("with_second_level.yaml", "bob.yaml")).NotTo(HaveOccurred())
			Expect(os.Rename("with_second_level_second_level.yaml", dir+"/second_level/bob.yaml")).NotTo(HaveOccurred())
			capture()

			ctx := context.Background()
			b, err := bob.Bob(bob.WithDir(dir), bob.WithCachingEnabled(false))
			Expect(err).NotTo(HaveOccurred())
			err = b.Build(ctx, "second_level/run-hello-second")
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("go version go1.17"))
		})

		It("running task go version from second level directory will use go_1_17 second level task dependencies", func() {
			b, err := bob.Bob(bob.WithDir(dir+"/second_level"), bob.WithCachingEnabled(false))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chdir(dir + "/second_level")
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "run-hello-second")
			Expect(err).NotTo(HaveOccurred())

			output := output()

			Expect(output).To(ContainSubstring("go version go1.17"))

			err = os.Chdir(dir)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with a task which depends on a second task", func() {
		It("will build dependencies of both tasks", func() {
			Expect(os.Rename("with_depends_on_dependency.yaml", "bob.yaml")).NotTo(HaveOccurred())
			Expect(os.Rename("with_depends_on_dependency_second_level.yaml", dir+"/second_level/bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			output := output()

			//run both tasks. `build` task can use `go` because it's a dependency of its dependson task
			Expect(output).To(ContainSubstring("Hello second!"))
			Expect(output).To(ContainSubstring("go version go1.16.15"))
		})
	})

	Context("with use_nix false in a second level bobfile", func() {
		It("will build dependencies of both tasks", func() {
			Expect(os.Rename("with_second_level_use_nix_false.yaml", "bob.yaml")).NotTo(HaveOccurred())
			Expect(os.Rename("with_second_level_use_nix_false_second_level.yaml", dir+"/second_level/bob.yaml")).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err := b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())
			output := output()

			// will run both because build depends on second
			Expect(output).To(ContainSubstring("Hello second!"))
			Expect(output).To(ContainSubstring("Hello build!"))

			// but should not build dependencies from second level because of use_nix false
			Expect(output).To(Not(ContainSubstring("go-1.16")))
		})
	})
})
