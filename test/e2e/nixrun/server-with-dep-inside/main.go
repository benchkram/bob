package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// This binary has a dependency on php
func main() {
	cmd := exec.Command("php", "-v")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	fmt.Println(out.String())
}
