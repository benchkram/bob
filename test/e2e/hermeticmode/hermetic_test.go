package hermeticmodetest

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	runTaskServer = "server"

	projectServer        = "server"
	projectServerWithEnv = "server-with-env"
)

/**
To get the list of current environment variables for specific tasks or binaries
we output the env command output in the ./envOutput file. In bobfiles we do that with
`env > envOutput` shell command, and in binaries we write the output of `env` command to the same envOutput file
*/
var _ = Describe("Testing hermetic mode for build tasks", func() {

	AfterEach(func() {
		err := os.Remove("./envOutput")
		Expect(err).To(BeNil())
	})

	Context("with default hermetic mode", func() {
		It("should have only 2 variables", func() {
			useBobfile("build_with_use_nix_true")
			defer releaseBobfile("build_with_use_nix_true")

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			envVariables, err := readLines("./envOutput")
			Expect(err).NotTo(HaveOccurred())

			// only HOME && PATH
			Expect(len(envVariables)).Should(Equal(2))
		})
	})

	Context("with --env VAR_ONE=somevalue", func() {
		It("should have 3 variables", func() {
			useBobfile("build_with_use_nix_true")
			defer releaseBobfile("build_with_use_nix_true")

			b, err := BobSetup(
				"VAR_ONE=somevalue",
			)
			Expect(err).NotTo(HaveOccurred())

			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			envVariables, err := readLines("./envOutput")
			Expect(err).NotTo(HaveOccurred())

			// will contain HOME && PATH && VAR_ONE
			Expect(len(envVariables)).Should(Equal(3))
			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())
		})
	})

	Context("with --env HOME=newHomeValue", func() {
		It("should have 2 variables and whitelisted HOME will be overwritten", func() {
			useBobfile("build_with_use_nix_true")
			defer releaseBobfile("build_with_use_nix_true")

			b, err := BobSetup(
				"HOME=newHomeValue",
			)
			Expect(err).NotTo(HaveOccurred())

			err = b.Build(ctx, "build")
			Expect(err).NotTo(HaveOccurred())

			envVariables, err := readLines("./envOutput")
			Expect(err).NotTo(HaveOccurred())

			// will contain HOME && PATH && VAR_ONE
			Expect(len(envVariables)).Should(Equal(2))
			// will overwrite whitelisted HOME
			Expect(keyHasValue("HOME", "newHomeValue", envVariables)).To(BeTrue())
		})
	})
})

var _ = Describe("Testing hermetic mode for init", func() {
	AfterEach(func() {
		err := os.Remove("./envOutput")
		Expect(err).To(BeNil())
	})

	Context("with default hermetic mode", func() {
		It("should have only 2 variables", func() {
			useBobfile("init_with_use_nix_true")
			defer releaseBobfile("init_with_use_nix_true")

			useProject(projectServer)
			defer releaseProject(projectServer)

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			var envVariables []string
			Eventually(func() error {
				envVariables, err = readLines("./envOutput")
				return err
			}, "5s").Should(BeNil())

			// only HOME && PATH
			Expect(len(envVariables)).Should(Equal(2))

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with --env VAR_ONE=somevalue", func() {
		It("should have 3 variables", func() {
			useBobfile("init_with_use_nix_true")
			defer releaseBobfile("init_with_use_nix_true")

			useProject(projectServer)
			defer releaseProject(projectServer)

			b, err := BobSetup(
				"VAR_ONE=somevalue",
			)
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			// will contain HOME && PATH && VAR_ONE
			var envVariables []string
			Eventually(func() int {
				envVariables, _ = readLines("./envOutput")
				return len(envVariables)
			}, "5s").Should(Equal(3))

			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Testing hermetic mode for initOnce", func() {

	AfterEach(func() {
		err := os.Remove("./envOutput")
		Expect(err).To(BeNil())
	})

	Context("with default hermetic mode", func() {
		It("should have only 2 variables", func() {
			useBobfile("init_once_with_use_nix_true")
			defer releaseBobfile("init_once_with_use_nix_true")

			useProject(projectServer)
			defer releaseProject(projectServer)

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			// only HOME && PATH
			Eventually(func() int {
				envVariables, _ := readLines("./envOutput")
				return len(envVariables)
			}, "5s").Should(Equal(2))

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with --env VAR_ONE=somevalue", func() {
		It("should have 3 variables", func() {
			useBobfile("init_once_with_use_nix_true")
			defer releaseBobfile("init_once_with_use_nix_true")

			useProject(projectServer)
			defer releaseProject(projectServer)

			b, err := BobSetup(
				"VAR_ONE=somevalue",
			)
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			// will contain HOME && PATH && VAR_ONE
			var envVariables []string
			Eventually(func() int {
				envVariables, _ = readLines("./envOutput")
				return len(envVariables)
			}, "5s").Should(Equal(3))

			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Testing hermetic mode for server", func() {
	AfterEach(func() {
		err := os.Remove("./envOutput")
		Expect(err).To(BeNil())
	})

	Context("with default hermetic mode", func() {
		It("should have only 2 variables", func() {
			useBobfile("binary_with_use_nix_true")
			defer releaseBobfile("binary_with_use_nix_true")

			useProject(projectServerWithEnv)
			defer releaseProject(projectServerWithEnv)

			b, err := BobSetup()
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			var envVariables []string
			Eventually(func() error {
				envVariables, err = readLines("./envOutput")
				return err
			}, "5s").Should(BeNil())

			// only HOME && PATH
			Expect(len(envVariables)).Should(Equal(2))

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with  --env VAR_ONE=somevalue", func() {
		It("should have 3 variables", func() {
			useBobfile("binary_with_use_nix_true")
			defer releaseBobfile("binary_with_use_nix_true")

			useProject(projectServerWithEnv)
			defer releaseProject(projectServerWithEnv)

			b, err := BobSetup(
				"VAR_ONE=somevalue",
			)
			Expect(err).NotTo(HaveOccurred())

			cmdr, err := b.Run(ctx, runTaskServer)
			Expect(err).NotTo(HaveOccurred())

			err = cmdr.Start()
			Expect(err).NotTo(HaveOccurred())

			var envVariables []string
			Eventually(func() error {
				envVariables, err = readLines("./envOutput")
				return err
			}, "5s").Should(BeNil())

			// will contain HOME && PATH && VAR_ONE
			Expect(len(envVariables)).Should(Equal(3))

			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

			err = cmdr.Stop()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
