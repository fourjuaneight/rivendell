package utils

import (
	"testing"
)

func TestFileNameFmt(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"spaces to underscores", "Hello World", "Hello_World"},
		{"dash separator", "Hello - World", "Hello-World"},
		{"colon separator", "Hello: World", "Hello-World"},
		{"double colon separator", "A :: B", "A-B"},
		{"em dash separator", "A — B", "A-B"},
		{"ampersand", "A & B", "A_and_B"},
		{"ampersand no spaces", "A&B", "A_and_B"},
		{"trailing period stripped", "Hello.", "Hello"},
		{"trailing question stripped", "Hello?", "Hello"},
		{"trailing exclamation stripped", "Hello!", "Hello"},
		{"emoji removed", "Hello 🎉", "Hello"},
		{"leading emoji removed", "🎉 Hello", "Hello"},
		{"special chars stripped", "Hello@World", "HelloWorld"},
		{"multiple spaces", "A  B", "A__B"},
		{"ellipsis separator", "A… B", "A_B"},
		{"pipe normalized to dash", "A|B", "A-B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileNameFmt(tt.input)
			if got != tt.want {
				t.Errorf("FileNameFmt(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCapitalized(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"hello world", "Hello World"},
		{"articles", "Articles"},
		{"already Capitalized", "Already Capitalized"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToCapitalized(tt.input)
			if got != tt.want {
				t.Errorf("ToCapitalized(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEmojiUnicode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no emoji passthrough", "hello world", "hello world"},
		{"emoji replaced", "hello 🎉", "hello U+1F389"},
		{"multiple emojis", "🎉🎊", "U+1F389U+1F38A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EmojiUnicode(tt.input)
			if got != tt.want {
				t.Errorf("EmojiUnicode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		typeStr  string
		url      string
		wantFile string
		wantMIME string
	}{
		{"articles", "", "md", "text/markdown"},
		{"podcasts", "", "mp3", "audio/mpeg"},
		{"videos", "", "mp4", "video/mp4"},
		{"comics", "https://example.com/cover.png", "png", "image/png"},
		{"comics", "https://example.com/cover.jpg", "jpg", "image/jpg"},
		{"comics", "https://example.com/cover.webp", "webp", "image/webp"},
	}

	for _, tt := range tests {
		t.Run(tt.typeStr+"_"+tt.wantFile, func(t *testing.T) {
			got := GetFileType(tt.typeStr, tt.url)
			if got.File != tt.wantFile {
				t.Errorf("GetFileType(%q, %q).File = %q, want %q", tt.typeStr, tt.url, got.File, tt.wantFile)
			}
			if got.MIME != tt.wantMIME {
				t.Errorf("GetFileType(%q, %q).MIME = %q, want %q", tt.typeStr, tt.url, got.MIME, tt.wantMIME)
			}
		})
	}
}
