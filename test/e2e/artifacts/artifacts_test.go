package artifactstest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/playbook"
	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/errz"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test artifact and target invalidation", func() {
	Context("in a fresh playground", func() {

		It("should initialize bob playground", func() {
			Expect(bob.CreatePlayground(dir)).NotTo(HaveOccurred())
		})

		It("should build", func() {
			ctx := context.Background()
			err := b.Build(ctx, "build")
			errz.Log(err)
			Expect(err).NotTo(HaveOccurred())
		})

		var artifactID string
		It("should check that exactly one artifact was created", func() {
			fs, err := os.ReadDir(artifactDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(fs)).To(Equal(1))
			artifactID = fs[0].Name()
		})

		// 5)
		It("should not rebuild but update the target", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 6)
		It("should not rebuild but unpack from local artifact", func() {
			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))
		})

		// 7)
		It("should rebuild and update local artifact", func() {
			err := artifactRemove(artifactID)
			Expect(err).NotTo(HaveOccurred())
			err = targetChange(filepath.Join(dir, "run"))
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateCompleted))

			exists, err := artifactExists(artifactID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		// 8)
		It("should update target from local artifact", func() {
			err := targetChange(filepath.Join(dir, "run"))
			Expect(err).NotTo(HaveOccurred())

			state, err := buildTask(b, "build")
			Expect(err).NotTo(HaveOccurred())
			Expect(state.State()).To(Equal(playbook.StateNoRebuildRequired))

			file.Exists(filepath.Join(dir, "run"))
		})

		It("cleanup", func() {
			err := b.Clean()
			Expect(err).NotTo(HaveOccurred())

			err = artifactsClean()
			Expect(err).NotTo(HaveOccurred())

			err = reset()
			Expect(err).NotTo(HaveOccurred())
		})

	})
})

// artifactRemove a artifact from the local artifact store
func artifactRemove(id string) error {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.Name() == id {
			err = os.Remove(filepath.Join(artifactDir, f.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// artifactExists checks if a artifact exists in the local artifact store
func artifactExists(id string) (exist bool, _ error) {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return false, err
	}

	for _, f := range fs {
		if f.Name() == id {
			exist = true
			break
		}
	}

	return exist, nil
}

// artifactsClean deletes all artifacts from the store
func artifactsClean() error {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		err = os.Remove(filepath.Join(artifactDir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// targetChanged appends a string to a target
func targetChange(dir string) error {
	f, err := os.OpenFile(dir, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString("change_the_target")
	if err != nil {
		return err
	}
	return f.Close()
}

// buildTask and returns it's state
func buildTask(b *bob.B, taskname string) (_ *playbook.Status, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)
	pb, err := aggregate.Playbook(taskname)
	errz.Fatal(err)

	err = pb.Build(context.Background())
	errz.Fatal(err)

	return pb.TaskStatus(taskname)
}
