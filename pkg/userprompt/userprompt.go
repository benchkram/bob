package userprompt

import (
	"bufio"
	"fmt"
	"github.com/benchkram/errz"
	"os"
	"strings"
)

func Confirm() (_ bool, err error) {
	defer errz.Recover(&err)
	writeMsg := func() { fmt.Println("Confirm [Y/n] ?") }
	writeMsg()
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		text = strings.TrimRight(text, "\n")
		errz.Fatal(err)
		if strings.ToLower(text) == "y" || text == "" {
			return true, nil
		} else if strings.ToLower(text) == "n" {
			return false, nil
		} else {
			writeMsg()
		}
	}
}
