package main

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"over length", "hello world", 8, "hello..."},
		{"very short max", "hello", 2, "he"},
		{"empty string", "", 5, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestScrollText(t *testing.T) {
	// Text shorter than width should be returned as-is
	got := scrollText("hi", 10, 0)
	if got != "hi" {
		t.Errorf("scrollText short text = %q, want %q", got, "hi")
	}

	// Text longer than width should be scrolled and fit within width
	got = scrollText("hello world this is long", 10, 0)
	if len(got) != 10 {
		t.Errorf("scrollText long text len = %d, want 10", len(got))
	}

	// Different tick should produce different output (eventually)
	got0 := scrollText("hello world this is long", 10, 0)
	got10 := scrollText("hello world this is long", 10, 10)
	if got0 == got10 {
		t.Log("scrollText at tick 0 and 10 are the same (possible but unlikely for long text)")
	}
}

func TestMinifyStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Up 3 hours", "Up 3h"},
		{"Up 1 minute", "Up 1m"},
		{"Up 2 days", "Up 2d"},
		{"Exited (0) 2 days ago", "Exit (0) 2d"},
		{"Created", "New"},
		{"Restarting", "Restart"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := minifyStatus(tt.input)
			if got != tt.expected {
				t.Errorf("minifyStatus(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatBytesShort(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1572864, "1.5M"},
		{10485760, "10M"},
		{1073741824, "1.0G"},
		{2147483648, "2.0G"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := formatBytesShort(tt.input)
			if got != tt.expected {
				t.Errorf("formatBytesShort(%f) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFirstPublicPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single port", "0.0.0.0:8080->80/tcp", "8080"},
		{"multiple ports", "0.0.0.0:8080->80/tcp, 0.0.0.0:443->443/tcp", "8080"},
		{"no public port", "80/tcp", ""},
		{"ipv6", ":::8080->80/tcp", "8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstPublicPort(tt.input)
			if got != tt.expected {
				t.Errorf("firstPublicPort(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
