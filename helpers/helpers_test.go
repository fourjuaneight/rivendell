package helpers

import (
	"testing"
)

func TestParseGHURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "standard https URL",
			url:       "https://github.com/fourjuaneight/rivendell",
			wantOwner: "fourjuaneight",
			wantRepo:  "rivendell",
		},
		{
			name:      "URL with trailing slash",
			url:       "https://github.com/owner/repo/",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "URL with path suffix",
			url:       "https://github.com/owner/repo/blob/main/README.md",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "non-github URL",
			url:     "https://example.com/foo",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseGHURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGHURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestParseMTGURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantErr bool
	}{
		{
			name:   "valid scryfall oembed URL",
			url:    "https://scryfall.com/cards/a3c99fd2-f037-4c3c-a07d-4b8e41b8bde0/oembed",
			wantID: "a3c99fd2-f037-4c3c-a07d-4b8e41b8bde0",
		},
		{
			name:    "URL without oembed path",
			url:     "https://scryfall.com/cards/abc123",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMTGURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMTGURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantID {
				t.Errorf("parseMTGURL(%q) = %q, want %q", tt.url, got, tt.wantID)
			}
		})
	}
}

func TestParseMDURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  string
		wantErr bool
	}{
		{
			name:   "standard mangadex title URL",
			url:    "https://mangadex.org/title/a96676be-9e5d-4d1f-88b0-d48ead35c978",
			wantID: "a96676be-9e5d-4d1f-88b0-d48ead35c978",
		},
		{
			name:   "URL with title slug",
			url:    "https://mangadex.org/title/a96676be-9e5d-4d1f-88b0-d48ead35c978/berserk",
			wantID: "a96676be-9e5d-4d1f-88b0-d48ead35c978",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMDURL(tt.url)
			if err != nil {
				t.Errorf("parseMDURL(%q) unexpected error: %v", tt.url, err)
				return
			}
			if got != tt.wantID {
				t.Errorf("parseMDURL(%q) = %q, want %q", tt.url, got, tt.wantID)
			}
		})
	}
}

func TestParseTMDBURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantID       string
		wantCategory string
		wantErr      bool
	}{
		{
			name:         "movie URL",
			url:          "https://www.themoviedb.org/movie/550-fight-club",
			wantID:       "550",
			wantCategory: "movie",
		},
		{
			name:         "TV show URL",
			url:          "https://www.themoviedb.org/tv/1396-breaking-bad",
			wantID:       "1396",
			wantCategory: "tv",
		},
		{
			name:         "URL without slug",
			url:          "https://www.themoviedb.org/movie/550",
			wantID:       "550",
			wantCategory: "movie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTMDBURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTMDBURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.id != tt.wantID {
					t.Errorf("id = %q, want %q", got.id, tt.wantID)
				}
				if got.category != tt.wantCategory {
					t.Errorf("category = %q, want %q", got.category, tt.wantCategory)
				}
			}
		})
	}
}

func TestCleanYTURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantID   string
	}{
		{
			name:   "short youtu.be URL",
			url:    "https://youtu.be/dQw4w9WgXcQ",
			wantID: "dQw4w9WgXcQ",
		},
		{
			name:   "full youtube.com watch URL",
			url:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantID: "dQw4w9WgXcQ",
		},
		{
			name:   "youtube.com without www",
			url:    "https://youtube.com/watch?v=dQw4w9WgXcQ",
			wantID: "dQw4w9WgXcQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanYTURL(tt.url)
			if got.Link != "https://youtu.be/"+tt.wantID {
				t.Errorf("cleanYTURL(%q).Link = %q, want %q", tt.url, got.Link, "https://youtu.be/"+tt.wantID)
			}
		})
	}
}

func TestEscapeText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"no newlines", "no newlines"},
		{"line\nbreak", `line\nbreak`},
		{"multi\nline\ntext", `multi\nline\ntext`},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeText(tt.input)
			if got != tt.want {
				t.Errorf("escapeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDiscogsTitle(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		wantArtist string
		wantAlbum  string
	}{
		{
			name:       "standard artist - album format",
			title:      "Radiohead - OK Computer",
			wantArtist: "Radiohead",
			wantAlbum:  "OK Computer",
		},
		{
			name:       "artist with dash in name",
			title:      "Nine Inch Nails - The Downward Spiral",
			wantArtist: "Nine Inch Nails",
			wantAlbum:  "The Downward Spiral",
		},
		{
			name:       "album with dash in title",
			title:      "Miles Davis - Kind Of Blue - Legacy Edition",
			wantArtist: "Miles Davis",
			wantAlbum:  "Kind Of Blue - Legacy Edition",
		},
		{
			name:       "no separator returns empty artist",
			title:      "SomeAlbumWithNoArtist",
			wantArtist: "",
			wantAlbum:  "SomeAlbumWithNoArtist",
		},
		{
			name:       "empty string",
			title:      "",
			wantArtist: "",
			wantAlbum:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArtist, gotAlbum := parseDiscogsTitle(tt.title)
			if gotArtist != tt.wantArtist {
				t.Errorf("artist = %q, want %q", gotArtist, tt.wantArtist)
			}
			if gotAlbum != tt.wantAlbum {
				t.Errorf("album = %q, want %q", gotAlbum, tt.wantAlbum)
			}
		})
	}
}
