package nix_test

import (
	"context"
	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("Testing new nix implementation", func() {
	BeforeEach(func() {
		Expect(os.Setenv("PATH", initialPath)).NotTo(HaveOccurred())
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

			Expect(output()).To(ContainSubstring("go version go1.17.8"))
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
})
