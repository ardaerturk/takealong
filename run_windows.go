package main

import (
	"os"
	"os/exec"
)

func execute(path string, args []string, dir string) error {
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
