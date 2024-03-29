package hermeticmodetest

// import (
// 	"os"

// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )

// const (
// 	runTaskServer = "server"

// 	projectServer        = "server"
// 	projectServerWithEnv = "server-with-env"
// )

// /*
// *
// To get the list of current environment variables for specific tasks or binaries
// we output the env command output in the ./envOutput file. In bobfiles we do that with
// `env > envOutput` shell command, and in binaries we write the output of `env` command to the same envOutput file
// */
// var _ = Describe("Testing hermetic mode for build tasks", func() {
// 	AfterEach(func() {
// 		err := os.Remove("./envOutput")
// 		Expect(err).To(BeNil())
// 	})

// 	Context("with default hermetic mode", func() {
// 		It("should have 79 variables", func() {
// 			useBobfile("build")
// 			defer releaseBobfile("build")

// 			b, err := BobSetup()
// 			Expect(err).NotTo(HaveOccurred())

// 			err = b.Build(ctx, "build")
// 			Expect(err).NotTo(HaveOccurred())

// 			envVariables, err := readLines("./envOutput")
// 			Expect(err).NotTo(HaveOccurred())

// 			// Always same environment of 79 variables
// 			Expect(len(envVariables)).Should(Equal(79))
// 		})
// 	})
// })

// var _ = Describe("Testing hermetic mode for init", func() {
// 	AfterEach(func() {
// 		err := os.Remove("./envOutput")
// 		Expect(err).To(BeNil())
// 	})

// 	Context("with default hermetic mode", func() {
// 		It("should have 78 variables", func() {
// 			useBobfile("init")
// 			defer releaseBobfile("init")

// 			useProject(projectServer)
// 			defer releaseProject(projectServer)

// 			b, err := BobSetup()
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			var envVariables []string
// 			Eventually(func() error {
// 				envVariables, err = readLines("./envOutput")
// 				return err
// 			}, "5s").Should(BeNil())
// 			Skip("todo: sometimes is 78 sometimes is 79")
// 			// Always same environment of 78 variables
// 			Expect(len(envVariables)).Should(Equal(78))

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})

// 	Context("with --env VAR_ONE=somevalue", func() {
// 		It("should have 80 variables", func() {
// 			useBobfile("init")
// 			defer releaseBobfile("init")

// 			useProject(projectServer)
// 			defer releaseProject(projectServer)

// 			b, err := BobSetup(
// 				"VAR_ONE=somevalue",
// 			)
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			// will contain 79 + VAR_ONE
// 			var envVariables []string
// 			Eventually(func() int {
// 				envVariables, _ = readLines("./envOutput")
// 				return len(envVariables)
// 			}, "5s").Should(Equal(80))

// 			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})
// })

// var _ = Describe("Testing hermetic mode for initOnce", func() {
// 	AfterEach(func() {
// 		err := os.Remove("./envOutput")
// 		Expect(err).To(BeNil())
// 	})

// 	Context("with default hermetic mode", func() {
// 		It("should have 79 variables", func() {
// 			useBobfile("init_once")
// 			defer releaseBobfile("init_once")

// 			useProject(projectServer)
// 			defer releaseProject(projectServer)

// 			b, err := BobSetup()
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			// Always same environment of 79 variables
// 			Eventually(func() int {
// 				envVariables, _ := readLines("./envOutput")
// 				return len(envVariables)
// 			}, "5s").Should(Equal(79))

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})

// 	Context("with --env VAR_ONE=somevalue", func() {
// 		It("should have 80 variables", func() {
// 			useBobfile("init_once")
// 			defer releaseBobfile("init_once")

// 			useProject(projectServer)
// 			defer releaseProject(projectServer)

// 			b, err := BobSetup(
// 				"VAR_ONE=somevalue",
// 			)
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			// will contain 79 vars + VAR_ONE
// 			var envVariables []string
// 			Eventually(func() int {
// 				envVariables, _ = readLines("./envOutput")
// 				return len(envVariables)
// 			}, "5s").Should(Equal(80))

// 			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})
// })

// var _ = Describe("Testing hermetic mode for server", func() {
// 	AfterEach(func() {
// 		err := os.Remove("./envOutput")
// 		Expect(err).To(BeNil())
// 	})

// 	Context("with default hermetic mode", func() {
// 		It("should have only 2 variables", func() {
// 			useBobfile("binary")
// 			defer releaseBobfile("binary")

// 			useProject(projectServerWithEnv)
// 			defer releaseProject(projectServerWithEnv)

// 			b, err := BobSetup()
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			var envVariables []string
// 			Eventually(func() error {
// 				envVariables, err = readLines("./envOutput")
// 				return err
// 			}, "5s").Should(BeNil())

// 			// only HOME && PATH
// 			Expect(len(envVariables)).Should(Equal(2))

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})

// 	Context("with  --env VAR_ONE=somevalue", func() {
// 		It("should have 3 variables", func() {
// 			useBobfile("binary")
// 			defer releaseBobfile("binary")

// 			useProject(projectServerWithEnv)
// 			defer releaseProject(projectServerWithEnv)

// 			b, err := BobSetup(
// 				"VAR_ONE=somevalue",
// 			)
// 			Expect(err).NotTo(HaveOccurred())

// 			cmdr, err := b.Run(ctx, runTaskServer)
// 			Expect(err).NotTo(HaveOccurred())

// 			err = cmdr.Start()
// 			Expect(err).NotTo(HaveOccurred())

// 			var envVariables []string
// 			Eventually(func() error {
// 				envVariables, err = readLines("./envOutput")
// 				return err
// 			}, "5s").Should(BeNil())

// 			// will contain HOME && PATH && VAR_ONE
// 			Expect(len(envVariables)).Should(Equal(3))

// 			Expect(keyHasValue("VAR_ONE", "somevalue", envVariables)).To(BeTrue())

// 			err = cmdr.Stop()
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})
// })
