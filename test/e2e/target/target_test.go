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
			buildinfos, err := buildInfoStore.GetBuildInfos()
			Expect(err).NotTo(HaveOccurred())

			var found bool
			for _, bi := range buildinfos {
				if bi.Info.Taskname == "all" {
					all = bi
					found = true
					break
				}
			}
			Expect(found).To(BeTrue())
			Expect(all).NotTo(BeNil())
		})

		It("target hashes of child tasks WITH a valid target must exist ", func() {
			Expect(len(all.Targets)).To(Equal(3))
		})

		It("target hashes of child tasks WITHOUT a valid target must NOT exist ", func() {
			// print does not have a target and therfore should not store a hash
			_, ok := all.Targets["second-level/third-level/print"]
			Expect(ok).To(BeFalse())
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

			buildinfo, err := buildInfoStore.GetBuildInfo(hashIn.String())
			Expect(err).NotTo(HaveOccurred())
			Expect(buildinfo).NotTo(BeNil())
			Expect(len(buildinfo.Targets)).To(Equal(1))
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

			buildinfo, err := buildInfoStore.GetBuildInfo(hashIn.String())
			Expect(err).NotTo(HaveOccurred())
			Expect(buildinfo).NotTo(BeNil())
			Expect(len(buildinfo.Targets)).To(Equal(0))
		})
	})
})
