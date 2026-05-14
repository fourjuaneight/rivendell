package utils

import "fmt"

// Download YT video from url.
func YTDL(url string, name string) error {
	if err := CMD("yt-dlp", "-f", "b", "--merge-output-format", "mp4", "-o", name, url); err != nil {
		return fmt.Errorf("[YTDL][yt-dlp]: %w", err)
	}
	return nil
}
