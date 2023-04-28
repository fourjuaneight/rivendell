package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// Execute a command and print its output to stderr.
func CMD(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("[CMD][cmd.StderrPipe]: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[CMD][cmd.Start]: %w", err)
	}

	buf := make([]byte, 1024)
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			os.Stderr.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("[CMD][cmd.Wait]: %w", err)
	}

	return nil
}
