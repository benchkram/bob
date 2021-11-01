package targettest

import (
	"context"
	"os"
	"time"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/bobtask/hash"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob's file target handling", func() {
	Context("in a fresh environment", func() {
		var err error

		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("runs build all", func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			err = b.Build(ctx, bob.BuildAllTargetName)
			Expect(err).NotTo(HaveOccurred())
			cancel()
		})

		var hashes hash.Hashes
		var all hash.Task
		It("hashes of task `all` must exist and be valid", func() {
			var ok bool

			task := bobtask.Make()
			hashes, err = task.ReadHashes()
			Expect(err).NotTo(HaveOccurred())

			all, ok = hashes["all"]
			Expect(ok).To(BeTrue())

		})

		It("target hashes of child tasks WITH a valid target must exist ", func() {

			targetall, ok := all.Targets["all"]
			Expect(ok).To(BeTrue())
			Expect(targetall).NotTo(BeEmpty())

			secondLevelBuild, ok := all.Targets["second-level/build2"]
			Expect(ok).To(BeTrue())
			Expect(secondLevelBuild).NotTo(BeEmpty())

			thirdLevelBuild, ok := all.Targets["second-level/third-level/build3"]
			Expect(ok).To(BeTrue())
			Expect(thirdLevelBuild).NotTo(BeEmpty())
		})

		It("target hashes of child tasks WITHOUT a valid target must NOT exist ", func() {
			// print does not have a target and therfore should not store a hash
			_, ok := all.Targets["second-level/third-level/print"]
			Expect(ok).To(BeFalse())
		})

		// ----- Check creation of hashes on child tasks -----

		It("target hash of task `/second-level/third-level/buil3` must exist", func() {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			_ = os.Chdir("./second-level/third-level")
			defer func() { _ = os.Chdir(wd) }()

			task := bobtask.Make()
			hashes, err = task.ReadHashes()
			Expect(err).NotTo(HaveOccurred())

			taskHash, ok := hashes["second-level/third-level/build3"]
			Expect(ok).To(BeTrue())
			Expect(len(taskHash.Targets)).To(Equal(1))
		})

		It("target hash of task `/second-level/third-level/print` must NOT exist", func() {
			wd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			_ = os.Chdir("./second-level/third-level")
			defer func() { _ = os.Chdir(wd) }()

			task := bobtask.Make()
			hashes, err = task.ReadHashes()
			Expect(err).NotTo(HaveOccurred())

			taskHash, ok := hashes["second-level/third-level/print"]
			Expect(ok).To(BeTrue())
			Expect(len(taskHash.Targets)).To(BeZero())
		})
	})
})
