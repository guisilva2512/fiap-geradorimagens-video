package processor

import (
	"context"
	"strings"
	"testing"
)

func TestFFmpegProcessReturnsCommandError(t *testing.T) {
	_, err := NewFFmpeg("ffmpeg-binary-does-not-exist", "1").Process(context.Background(), "input.mp4", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "ffmpeg:") {
		t.Fatalf("expected ffmpeg command error, got %v", err)
	}
}
