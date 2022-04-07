package exporttest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test bob's file export validation", func() {
	Context("in a fresh environment", func() {
		It("initializes bob playground", func() {
			Expect(bob.CreatePlayground(bob.PlaygroundOptions{Dir: dir})).NotTo(HaveOccurred())
		})

		It("run verify", func() {
			err := b.Verify(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})

		It("check that env vars are correctly set", func() {
			bobfile, err := b.Aggregate()
			Expect(err).NotTo(HaveOccurred())
			task, ok := bobfile.BTasks["generate"]
			Expect(ok).To(BeTrue())

			env := task.Env()
			Expect(env).To(ContainElement(ContainSubstring("OPENAPI_PROVIDER_PROJECT_OPENAPI_OPENAPI")))
			Expect(env).To(ContainElement(ContainSubstring("OPENAPI_PROVIDER_PROJECT_OPENAPI_OPENAPI2")))
			Expect(env).To(ContainElement(ContainSubstring("openapi-provider-project/openapi.yaml")))
			Expect(env).To(ContainElement(ContainSubstring("openapi-provider-project/openapi2.yaml")))
		})

		It("invalidate openapi provider export by deleting openapi.yaml file", func() {
			err := os.RemoveAll(filepath.Join(bob.SecondLevelOpenapiProviderDir, "openapi.yaml"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("run verify and expect it to fail", func() {
			err := b.Verify(context.Background())
			Expect(err).To(HaveOccurred())
		})
	})
})
