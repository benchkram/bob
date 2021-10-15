package execctl

import (
	"bufio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

var (
	// script that can be interrupted and does some cleanup if it is
	script = []byte(`
		cleanup() {
			echo "interrupted"
			sleep 1
			echo "exited"
			exit 0
		}
		
		trap cleanup INT 
		
		sleep .5
		
		echo "running"
		sleep 1
		echo "exited"
		exit 0
	`)
	// script that just errors
	scriptErr = []byte(`
		exit -1
	`)
	// script that prints the user-provided input to stderr
	scriptEchoErr = []byte(`
		read var
		echo $var >&2
		exit 0
	`)
)

var (
	tmpDir            string
	scriptPath        string
	scriptErrPath     string
	scriptEchoErrPath string
)

var _ = BeforeSuite(func() {
	var err error
	tmpDir, err = os.MkdirTemp("", "execctl-*")
	Expect(err).NotTo(HaveOccurred())

	scriptPath, err = createTempScript(tmpDir, script)
	Expect(err).NotTo(HaveOccurred())
	Expect(scriptPath).NotTo(BeEmpty())

	scriptErrPath, err = createTempScript(tmpDir, scriptErr)
	Expect(err).NotTo(HaveOccurred())
	Expect(scriptErrPath).NotTo(BeEmpty())

	scriptEchoErrPath, err = createTempScript(tmpDir, scriptEchoErr)
	Expect(err).NotTo(HaveOccurred())
	Expect(scriptErrPath).NotTo(BeEmpty())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(tmpDir)
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Test command start and wait", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command exited gracefully", func() {
		err := cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})
})

var _ = Describe("Test command stop", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command interrupted", func() {
		// allow the command to start running but don't give it enough time to exit gracefully
		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		err = cmd.Stop()
		Expect(err).NotTo(HaveOccurred())

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("interrupted"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})
})

var _ = Describe("Test command stop when already exited gracefully", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command was not interrupted (it was allowed to gracefully exit)", func() {
		// let the command exit gracefully
		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))

		err = cmd.Stop()
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Test command manual restart", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command interrupted", func() {
		// don't give the command enough time to exit gracefully
		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		err = cmd.Stop()
		Expect(err).NotTo(HaveOccurred())

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("interrupted"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})

	It("command started again and exited gracefully", func() {
		err := cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})
})

var _ = Describe("Test command restart", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command is restarted (interrupted and then started and exits gracefully)", func() {
		// interrupt the command with a restart
		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		err = cmd.Restart()
		Expect(err).NotTo(HaveOccurred())

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("interrupted"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))

		err = cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})
})

var _ = Describe("Test Wait() called multiple times on command that succeeded", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stdout())
	})

	It("command is awaited multiple times without errors", func() {
		err := cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("running"))

		l, err = readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal("exited"))
	})
})

var _ = Describe("Test Wait() called multiple times on command that returned error", func() {
	var cmd *Cmd

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptErrPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())
	})

	It("command awaited multiple times and returned an error on all", func() {
		err := cmd.Wait()
		Expect(err).To(HaveOccurred()) // original interrupt error

		err = cmd.Wait()
		Expect(err).To(HaveOccurred()) // from cmd.lastErr
	})
})

var _ = Describe("Test Stop() called multiple times on command that returned error", func() {
	var cmd *Cmd

	It("command started", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptErrPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())
	})

	It("command stopped multiple times and returned an error on all", func() {
		err := cmd.Wait()
		Expect(err).To(HaveOccurred()) // original exit error

		err = cmd.Stop()
		Expect(err).To(HaveOccurred()) // from cmd.lastErr

		err = cmd.Stop()
		Expect(err).To(HaveOccurred()) // from cmd.lastErr
	})
})

var _ = Describe("Test Start() called multiple times", func() {
	var cmd *Cmd

	It("command started multiple times", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptErrPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Test Stdin() and Stderr()", func() {
	var cmd *Cmd
	var r *bufio.Reader

	It("command started multiple times", func() {
		var err error
		cmd, err = NewCmd("test", "/bin/bash", "-c", scriptEchoErrPath)
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Start()
		Expect(err).NotTo(HaveOccurred())

		r = bufio.NewReader(cmd.Stderr())
	})

	It("command echoed user input to stderr", func() {
		in := "ping!"
		lf := "\n"

		_, err := cmd.Stdin().Write([]byte(in + lf))
		Expect(err).NotTo(HaveOccurred())

		err = cmd.Wait()
		Expect(err).NotTo(HaveOccurred())

		l, err := readLine(r)
		Expect(err).NotTo(HaveOccurred())
		Expect(l).To(Equal(in))
	})
})

func TestExecctl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "execctl suite")
}

func createTempScript(dir string, b []byte) (string, error) {
	f, err := os.CreateTemp(dir, "shell-*.sh")
	if err != nil {
		return "", err
	}

	path := f.Name()

	_, err = f.Write(b)
	if err != nil {
		return "", err
	}

	err = f.Chmod(0775)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	return path, nil
}

func readLine(r *bufio.Reader) (string, error) {
	l, _, err := r.ReadLine()
	if err != nil {
		return "", err
	}
	return string(l), nil
}
