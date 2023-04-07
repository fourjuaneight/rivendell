package utils

import (
	"log"
	"os"
)

// Delete list of files.
func DeleteFiles(files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Fatal("[DeleteFiles]: %w", err)
		}
	}
}
