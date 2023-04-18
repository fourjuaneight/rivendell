package utils

import (
	"fmt"
	"strings"
)

func installBin() error {
	wgetErr := CMD("sudo", "wget", "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp", "-O", "/usr/local/bin/yt-dlp")
	if wgetErr != nil {
		return fmt.Errorf("[installBin][wget]: %w", wgetErr)
	}

	chmodErr := CMD("sudo", "chmod", "a+rx", "/usr/local/bin/yt-dlp")
	if chmodErr != nil {
		return fmt.Errorf("[installBin][chmod]: %w", chmodErr)
	}

	return nil
}

// Download YT video from url.
func YTDL(url string, name string) error {
	err := CMD("yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/bestvideo+bestaudio:", "--merge-output-format", "mp4", "-o", name, url)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			err := installBin()
			if err != nil {
				return fmt.Errorf("[YTDL]: %w", err)
			}
		}

		return fmt.Errorf("[YTDL][yt-dlp]: %w", err)
	}

	return nil
}
