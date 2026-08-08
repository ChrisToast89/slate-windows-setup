package main

import "os/exec"

func openExe(path string) error {
	cmd := exec.Command(path)
	return cmd.Start()
}
