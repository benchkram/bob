package targettest

import (
	"context"
	"time"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob's file target handling", func() {
	Context("in a fresh environment", func() {
		var err error

		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		var hashInBeforeBuild string
		var hashInAfterBuild string
		It("should read input hash before build", func() {
			aggregate, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			globaltaskname := "second-level/third-level/build3"

			err = b.Nix().BuildNixDependenciesInPipeline(aggregate, globaltaskname)
			errz.Fatal(err)

			task, ok := aggregate.BTasks[globaltaskname]
			Expect(ok).To(BeTrue())

			hashIn, err := task.HashIn()
			Expect(err).NotTo(HaveOccurred())

			hashInBeforeBuild = hashIn.String()
		})

		It("runs build all", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			err = b.Build(ctx, bob.BuildAllTargetName)
			Expect(err).NotTo(HaveOccurred())
			cancel()
		})

		It("should read input hash after build", func() {
			aggregate, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			globaltaskname := "second-level/third-level/build3"
			err = b.Nix().BuildNixDependenciesInPipeline(aggregate, globaltaskname)
			errz.Fatal(err)

			task, ok := aggregate.BTasks[globaltaskname]
			Expect(ok).To(BeTrue())

			hashIn, err := task.HashIn()
			Expect(err).NotTo(HaveOccurred())

			hashInAfterBuild = hashIn.String()
		})

		It("input hash should be equal before and after build", func() {
			Expect(hashInAfterBuild).To(Equal(hashInBeforeBuild))
		})

		// var hashes hash.Hashes
		var all *buildinfo.I
		It("hashes of task `all` must exist and be valid", func() {
			buildinfos, err := buildinfoStore.GetBuildInfos()
			Expect(err).NotTo(HaveOccurred())

			var found bool
			for _, bi := range buildinfos {
				if bi.Meta.Task == "all" {
					all = bi
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
			Expect(all).NotTo(BeNil())
		})

		It("target checksum must be non empty", func() {
			Expect(all.Target.Filesystem.Hash).NotTo(BeEmpty())
		})

		// ----- Check creation of hashes on child tasks -----

		It("target hash of task `/second-level/third-level/build3` must exist and must contain one target", func() {
			aggregate, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			globaltaskname := "second-level/third-level/build3"
			err = b.Nix().BuildNixDependenciesInPipeline(aggregate, globaltaskname)
			errz.Fatal(err)

			task, ok := aggregate.BTasks[globaltaskname]
			Expect(ok).To(BeTrue())

			hashIn, err := task.HashIn()
			Expect(err).NotTo(HaveOccurred())

			buildinfo, err := buildinfoStore.GetBuildInfo(hashIn.String())
			Expect(err).NotTo(HaveOccurred())
			Expect(buildinfo).NotTo(BeNil())
			Expect(buildinfo.Target.Filesystem.Hash).NotTo(BeEmpty())
		})

		It("target hash of task `/second-level/third-level/print` must NOT exist", func() {
			aggregate, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())

			globaltaskname := "second-level/third-level/print"
			err = b.Nix().BuildNixDependenciesInPipeline(aggregate, globaltaskname)
			errz.Fatal(err)

			task, ok := aggregate.BTasks[globaltaskname]
			Expect(ok).To(BeTrue())

			hashIn, err := task.HashIn()
			Expect(err).NotTo(HaveOccurred())

			buildinfo, err := buildinfoStore.GetBuildInfo(hashIn.String())
			Expect(err).NotTo(HaveOccurred())
			Expect(buildinfo).NotTo(BeNil())
			Expect(buildinfo.Target.Filesystem.Hash).To(BeEmpty())
		})
	})
})
