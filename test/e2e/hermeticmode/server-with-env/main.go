package main

import (
	"bytes"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("env")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./envOutput", out.Bytes(), 0644)

	if err != nil {
		panic(err)
	}
}
