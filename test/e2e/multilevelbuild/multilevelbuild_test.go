package multilevelbuildtest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/playbook"
	"github.com/Benchkram/bob/pkg/file"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type binaryOutputFixture struct {
	path   string
	output string
}

type requiresRebuildFixture struct {
	taskname        string
	requiresRebuild bool
}

var _ = Describe("Test bob multilevel build", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("runs build all", func() {
			Expect(b.Build(context.Background(), bob.BuildAllTargetName)).NotTo(HaveOccurred())
		})

		binaries := []binaryOutputFixture{
			{
				path:   filepath.Join(dir, "run"),
				output: "Hello Playground v1\n",
			},
			{
				path:   filepath.Join(dir, bob.SecondLevelDir, "runsecondlevel"),
				output: "Hello Playground v2\n",
			},
			{
				path:   filepath.Join(dir, bob.SecondLevelDir, bob.ThirdLevelDir, "runthirdlevel"),
				output: "Hello Playground v3\n",
			},
		}

		It("checks that the built binaries exist", func() {
			for _, b := range binaries {
				Expect(file.Exists(b.path)).To(BeTrue(), fmt.Sprintf("%s doesn't exist", b.path))
			}
		})

		It("checks that the built binaries produce the expected output", func() {
			for _, b := range binaries {
				cmd := exec.Command("./" + b.path)
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())

				Expect(b.output).To(Equal(stderr.String()))
			}
		})

		It("runs build multilinetouch", func() {
			Expect(b.Build(context.Background(), "multilinetouch")).NotTo(HaveOccurred())
		})

		It("checks that the touched files really exist", func() {
			files := []string{
				"multilinefile1",
				"multilinefile2",
				"multilinefile3",
				"multilinefile4",
				"multilinefile5",
			}

			for _, f := range files {
				Expect(file.Exists(f)).To(BeTrue(), fmt.Sprintf("%s doesn't exist", f))
			}
		})

		It("checks that we do not require a rebuild of any of the levels", func() {
			fixtures := []requiresRebuildFixture{
				{
					taskname:        bob.BuildAllTargetName,
					requiresRebuild: false,
				},
				{
					taskname:        "second-level/build2",
					requiresRebuild: false,
				},
				{
					taskname:        "second-level/third-level/build3",
					requiresRebuild: false,
				},
			}

			requiresRebuildMustMatchFixtures(b, fixtures)
		})

		It("changes a file of the second-level", func() {
			f := filepath.Join(dir, bob.SecondLevelDir, "main2.go")
			c, err := ioutil.ReadFile(f)
			Expect(err).NotTo(HaveOccurred())

			c = append(c, []byte("// some random comment so the file content is changed")...)

			err = ioutil.WriteFile(f, c, 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("checks that we now require a rebuild of the second- and first-level, but not the third-level", func() {
			fixtures := []requiresRebuildFixture{
				{
					taskname:        bob.BuildAllTargetName,
					requiresRebuild: true,
				},
				{
					taskname:        "second-level/build2",
					requiresRebuild: true,
				},
				{
					taskname:        "second-level/third-level/build3",
					requiresRebuild: false,
				},
			}

			requiresRebuildMustMatchFixtures(b, fixtures)
		})

		It("runs build all again", func() {
			Expect(b.Build(context.Background(), bob.BuildAllTargetName)).NotTo(HaveOccurred())
		})

		It("checks that we do not require a rebuild of any of the levels", func() {
			fixtures := []requiresRebuildFixture{
				{
					taskname:        bob.BuildAllTargetName,
					requiresRebuild: false,
				},
				{
					taskname:        "second-level/build2",
					requiresRebuild: false,
				},
				{
					taskname:        "second-level/third-level/build3",
					requiresRebuild: false,
				},
			}

			requiresRebuildMustMatchFixtures(b, fixtures)
		})

		It("changes a file of the third-level", func() {
			f := filepath.Join(dir, bob.SecondLevelDir, bob.ThirdLevelDir, "main3.go")
			c, err := ioutil.ReadFile(f)
			Expect(err).NotTo(HaveOccurred())

			c = append(c, []byte("// some random comment so the file content is changed")...)

			err = ioutil.WriteFile(f, c, 0644)
			Expect(err).NotTo(HaveOccurred())
		})

		It("checks that we now require a rebuild of the third-, second- and first-level", func() {
			fixtures := []requiresRebuildFixture{
				{
					taskname:        bob.BuildAllTargetName,
					requiresRebuild: true,
				},
				{
					taskname:        "second-level/build2",
					requiresRebuild: true,
				},
				{
					taskname:        "second-level/third-level/build3",
					requiresRebuild: true,
				},
			}

			requiresRebuildMustMatchFixtures(b, fixtures)
		})
	})
})

func requiresRebuildMustMatchFixtures(b *bob.B, fixtures []requiresRebuildFixture) {
	aggregate, err := b.Aggregate()
	Expect(err).NotTo(HaveOccurred())
	pb, err := aggregate.Playbook(bob.BuildAllTargetName)
	Expect(err).NotTo(HaveOccurred())

	err = b.BuildTask(context.Background(), bob.BuildAllTargetName, pb)
	Expect(err).NotTo(HaveOccurred())

	for _, f := range fixtures {
		ts, err := pb.TaskStatus(f.taskname)
		Expect(err).NotTo(HaveOccurred())
		requiresRebuild := ts.State() != playbook.StateNoRebuildRequired

		Expect(f.requiresRebuild).To(Equal(requiresRebuild), fmt.Sprintf("task's %q rebuild requirement differ, got: %t, want: %t", f.taskname, requiresRebuild, f.requiresRebuild))
	}
}
