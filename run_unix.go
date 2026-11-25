//go:build !windows

package main

import (
	"os"
	"syscall"
)

func execute(path string, args []string, dir string) error {
	if dir != "" {
		if e := os.Chdir(dir); e != nil {
			return e
		}
	}
	return syscall.Exec(path, append([]string{path}, args...), os.Environ())
}
