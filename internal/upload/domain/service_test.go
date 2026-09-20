package domain

import "testing"

func TestIsValidVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		valid    bool
	}{
		{name: "mp4", filename: "video.mp4", valid: true},
		{name: "uppercase extension", filename: "video.MKV", valid: true},
		{name: "unsupported extension", filename: "video.txt", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidVideoFile(test.filename); got != test.valid {
				t.Fatalf("isValidVideoFile(%q) = %v, want %v", test.filename, got, test.valid)
			}
		})
	}
}
