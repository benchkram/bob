package nixruntest

import (
	"context"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/benchkram/bob/pkg/filepathutil"
)

var _ = Describe("Testing new nix implementation", func() {
	BeforeEach(func() {
		err := os.Setenv("PATH", initialPath)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		filepathutil.ClearListRecursiveCache()
	})
	Context("with use-nix false", func() {
		It("run without errors", func() {
			useBobfile("with_use_nix_false")
			defer releaseBobfile("with_use_nix_false")

			useProject("server")
			defer releaseProject("server")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			_, err = b.Run(ctx, "server")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("init with use-nix true and dependencies set on bob file", func() {
		It("installs dependencies set on bob file level for init", func() {
			useBobfile("init_with_bob_dependencies")
			defer releaseBobfile("init_with_bob_dependencies")

			useProject("server")
			defer releaseProject("server")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(500 * time.Millisecond) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("PHP 8.0.18"))
		})
	})

	Context("init with use-nix true and dependencies set on task level", func() {
		It("installs dependencies set on task level for init", func() {
			useBobfile("init_with_task_dependencies")
			defer releaseBobfile("init_with_task_dependencies")

			useProject("server")
			defer releaseProject("server")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(500 * time.Millisecond) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("PHP 8.0.18"))
		})
	})

	Context("initOnce with use-nix true and dependencies set on bob file", func() {
		It("installs dependencies set on bob file level for initOnce", func() {
			useBobfile("init_once_with_bob_dependencies")
			defer releaseBobfile("init_once_with_bob_dependencies")

			useProject("server")
			defer releaseProject("server")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(500 * time.Millisecond) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("PHP 8.0.18"))
		})
	})

	Context("initOnce with use-nix true and dependencies set on task level", func() {
		It("installs dependencies set on task level for initOnce", func() {
			useBobfile("init_once_with_task_dependencies")
			defer releaseBobfile("init_once_with_task_dependencies")

			useProject("server")
			defer releaseProject("server")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			capture()

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(500 * time.Millisecond) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
			Expect(output()).To(ContainSubstring("PHP 8.0.18"))
		})
	})

	Context("binary with bob dependencies", func() {
		It("installs dependencies set on bob file level", func() {
			// update bob.yaml with mock content
			err := os.Rename("binary_with_bob_dependencies.yaml", "bob.yaml")
			Expect(err).NotTo(HaveOccurred())

			useProject("server-with-dep-inside")
			defer releaseProject("server-with-dep-inside")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(1 * time.Second) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())

			// Reading the output of server tab should show php version
			for _, v := range cmdr.Subcommands() {
				if v.Name() != "server" {
					continue
				}
				buf := make([]byte, 30)
				_, err := io.ReadFull(v.Stdout(), buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(buf)).To(ContainSubstring("PHP 8.0.18"))
			}
		})
	})

	Context("binary with task dependencies", func() {
		It("installs dependencies set on task level", func() {
			useBobfile("binary_with_task_dependencies")
			defer releaseBobfile("binary_with_task_dependencies")

			useProject("server-with-dep-inside")
			defer releaseProject("server-with-dep-inside")

			b, err := Bob()
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(context.Background(), "server")
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(500 * time.Millisecond) // fixme

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())

			// Reading the output of server tab should show php version
			for _, v := range cmdr.Subcommands() {
				if v.Name() != "server" {
					continue
				}
				buf := make([]byte, 30)
				_, err := io.ReadFull(v.Stdout(), buf)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(buf)).To(ContainSubstring("PHP 8.0.18"))
			}
		})
	})
})
