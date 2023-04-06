package main

import (
	"log"
	"os"
	"os/exec"
)

// Download YT video from url.
func ytdl(url string, name string) {
	cmd := exec.Command("/usr/local/bin/yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/bestvideo+bestaudio:", "--merge-output-format", "mp4", "-o", name, url)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal("[ytdl][cmd.StderrPipe]: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal("[ytdl][cmd.Start]: %w", err)
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
		log.Fatal("[ytdl][cmd.Wait]: %w", err)
	}
}
