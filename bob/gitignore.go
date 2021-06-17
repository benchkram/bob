package bob

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/errz"
)

const gitignore = ".gitignore"

// func (b *B) gitignoreRemove(name string) (err error) {
// 	defer errz.Recover(&err)

// 	file, err := os.Open(gitignore)
// 	errz.Fatal(err)
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	// optionally, resize scanner's capacity for lines over 64K, see next example
// 	for scanner.Scan() {
// 		fmt.Println(scanner.Text())
// 	}

// 	return nil
// }

// gitignoreAdd a dir to the end of the .gitignore file.
func (b *B) gitignoreAdd(dir string) (err error) {
	defer errz.Recover(&err)

	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	file, err := os.OpenFile(filepath.Join(b.dir, gitignore), os.O_RDWR|os.O_CREATE, 0664)
	errz.Fatal(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text == dir {
			// Already on ignore list, no need to do anything
			return nil
		}
	}

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "") // Make sure to write to a new line
	fmt.Fprint(w, dir)
	w.Flush()

	return nil
}
