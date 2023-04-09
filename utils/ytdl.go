package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// Download YT video from url.
func YTDL(url string, name string) error {
	cmd := exec.Command("/usr/local/bin/yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/bestvideo+bestaudio:", "--merge-output-format", "mp4", "-o", name, url)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("[YTDL][cmd.StderrPipe]: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("[YTDL][cmd.Start]: %w", err)
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
		return fmt.Errorf("[YTDL][cmd.Wait]: %w", err)
	}

	return nil
}
