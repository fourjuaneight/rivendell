package datetime

import (
	"testing"
	"time"
)

func TestParseISO(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantYear int
		wantMonth time.Month
		wantDay  int
	}{
		{
			name:      "valid UTC offset",
			input:     "2024-03-15T10:30:00.000000+00:00",
			wantErr:   false,
			wantYear:  2024,
			wantMonth: time.March,
			wantDay:   15,
		},
		{
			name:      "valid negative offset",
			input:     "2024-03-15T10:30:00.000000-05:00",
			wantErr:   false,
			wantYear:  2024,
			wantMonth: time.March,
			wantDay:   15,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISO(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseISO(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Year() != tt.wantYear {
					t.Errorf("year = %d, want %d", got.Year(), tt.wantYear)
				}
				if got.Month() != tt.wantMonth {
					t.Errorf("month = %v, want %v", got.Month(), tt.wantMonth)
				}
				if got.Day() != tt.wantDay {
					t.Errorf("day = %d, want %d", got.Day(), tt.wantDay)
				}
			}
		})
	}
}

func TestSubDays(t *testing.T) {
	base := time.Date(2024, time.March, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		days    int
		wantDay int
	}{
		{0, 15},
		{1, 14},
		{7, 8},
		{15, 29}, // wraps to Feb 29 (2024 is leap year)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := SubDays(base, tt.days)
			if got.Day() != tt.wantDay {
				t.Errorf("SubDays(base, %d).Day() = %d, want %d", tt.days, got.Day(), tt.wantDay)
			}
		})
	}
}

func TestSubHours(t *testing.T) {
	base := time.Date(2024, time.March, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		hours    int
		wantHour int
		wantDay  int
	}{
		{0, 12, 15},
		{3, 9, 15},
		{12, 0, 15},
		{13, 23, 14}, // wraps to previous day
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := SubHours(base, tt.hours)
			if got.Hour() != tt.wantHour {
				t.Errorf("SubHours(base, %d).Hour() = %d, want %d", tt.hours, got.Hour(), tt.wantHour)
			}
			if got.Day() != tt.wantDay {
				t.Errorf("SubHours(base, %d).Day() = %d, want %d", tt.hours, got.Day(), tt.wantDay)
			}
		})
	}
}

func TestIsAfter(t *testing.T) {
	t1 := time.Date(2024, time.March, 15, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, time.March, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		date1 time.Time
		date2 time.Time
		want  bool
	}{
		{"later is after earlier", t1, t2, true},
		{"earlier is not after later", t2, t1, false},
		{"equal dates", t1, t1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAfter(tt.date1, tt.date2)
			if got != tt.want {
				t.Errorf("IsAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}
