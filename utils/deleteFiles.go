package main

import (
	"log"
	"os"
)

// Delete list of files.
func deleteFiles(files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Fatal("[deleteFiles]: %w", err)
		}
	}
}
