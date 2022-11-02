package nixtest

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/filepathutil"
	"github.com/benchkram/bob/pkg/nix"
)

var nixCacheContent = `5710bc1834e1cda5:/nix/store/sh0x9kihzkdz15x18ldg989pf29m4nm7-go-1.17.9
2ad93148f49750ec:/nix/store/6jpdfshhyqic7a85j02scrbwcxh2lfic-git-2.35.3
2954a404deb42501:/nix/store/crzgdqkbiaz30z709j8xbal7ylm29j42-go-1.18.1
e33686c0c937b507:/nix/store/h59dfk7dwrn7d2csykh9z9xm2miqmrnz-hello-2.12
`

var _ = Describe("Testing new nix implementation", func() {
	AfterEach(func() {
		filepathutil.ClearListRecursiveCache()
	})

	Context("with task dependencies", func() {
		It("task dependency go_1_17 will have priority to bob file go_1_18", func() {
			Expect(os.Rename("with_task_dependencies.yaml", "bob.yaml")).NotTo(HaveOccurred())

			nixCacheFilePath := dir + "customFile"
			defer os.Remove(nixCacheFilePath)

			fileCache, err := nix.NewCacheStore(nix.WithPath(nixCacheFilePath))
			Expect(err).NotTo(HaveOccurred())

			nixBuilder := bob.NewNixBuilder(bob.WithCache(fileCache))
			Expect(err).NotTo(HaveOccurred())

			b, err := bob.Bob(bob.WithDir(dir), bob.WithCachingEnabled(false), bob.WithNixBuilder(nixBuilder))
			Expect(err).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.17.9"))

			// build paths are cached in the custom file
			data, err := os.ReadFile(nixCacheFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(nixCacheContent))
		})
	})

	Context("with dependencies per bob file", func() {
		It("running task go version will use go_1_16 from bob file dependency", func() {
			Expect(os.Rename("with_bob_dependencies.yaml", "bob.yaml")).NotTo(HaveOccurred())

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.16.15"))
		})
	})

	Context("with ambiguous deps in root", func() {
		It("running task go version will use go_1_17 from bob file dependency", func() {
			Expect(os.Rename("with_ambiguous_deps_in_root.yaml", "bob.yaml")).NotTo(HaveOccurred())

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "run-hello")
			Expect(err).NotTo(HaveOccurred())

			Expect(output()).To(ContainSubstring("go version go1.17"))
		})
	})

	Context("with ambiguous deps in task", func() {
		It("running task go version will use go_1_17 from task deps", func() {
			Expect(os.Rename("with_ambiguous_deps_in_task.yaml", "bob.yaml")).NotTo(HaveOccurred())

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()
			ctx := context.Background()
			err = b.Build(ctx, "run-hello")
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
			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())
			err = b.Build(ctx, "second_level/run-hello-second")
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("go version go1.17"))
		})

		It("running task go version from second level directory will use go_1_17 second level task dependencies", func() {
			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			err = os.Chdir(dir + "/second_level")
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			capture()
			err = b.Build(ctx, "run-hello-second")
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("go version go1.17"))

			err = os.Chdir(dir)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
