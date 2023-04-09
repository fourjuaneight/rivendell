package utils

import (
	"fmt"
	"os"
)

// Delete list of files.
func DeleteFiles(files []string) error {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			return fmt.Errorf("[DeleteFiles]: %w", err)
		}
	}

	return nil
}
