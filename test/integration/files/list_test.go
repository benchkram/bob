package filestest

import (
	"path/filepath"

	"github.com/Benchkram/bob/pkg/filepathutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test file-related functions", func() {
	Context("in a fresh environment", func() {
		It("globs file", func() {
			fixtures := []struct {
				glob string
				want []string
			}{
				{
					glob: filepath.Join(dir, "**", "f_*_t*"),
					want: []string{
						"./files/deeplynested/a/b/f_b_two",
						"./files/deeplynested/a/b/c/f_c_two",
						"./files/deeplynested/a/b/c/d/f_d_two",
					},
				},
			}

			for _, f := range fixtures {
				got, err := filepathutil.ListRecursive(f.glob)
				Expect(err).NotTo(HaveOccurred())
				for i := range got {
					got[i] = stripBase(got[i])
				}
				want := f.want

				err = hasSameElements(got, want)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})
