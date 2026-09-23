package processor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv("FFMPEG_HELPER"); mode != "" {
		if mode == "success" {
			pattern := os.Args[len(os.Args)-1]
			frame := strings.Replace(pattern, "%06d", "000001", 1)
			_ = os.WriteFile(frame, []byte("frame"), 0o600)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestFFmpegProcessReturnsCommandError(t *testing.T) {
	_, err := NewFFmpeg("ffmpeg-binary-does-not-exist", "1").Process(context.Background(), "input.mp4", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "ffmpeg:") {
		t.Fatalf("expected ffmpeg command error, got %v", err)
	}
}

func TestFFmpegProcessReturnsGeneratedFrames(t *testing.T) {
	t.Setenv("FFMPEG_HELPER", "success")
	outputDir := t.TempDir()
	files, err := NewFFmpeg(os.Args[0], "1").Process(context.Background(), filepath.Join(outputDir, "input.mp4"), outputDir)
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one generated frame, got files=%v err=%v", files, err)
	}
}

func TestFFmpegProcessRejectsEmptyOutput(t *testing.T) {
	t.Setenv("FFMPEG_HELPER", "empty")
	_, err := NewFFmpeg(os.Args[0], "1").Process(context.Background(), "input.mp4", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "não gerou imagens") {
		t.Fatalf("expected empty output error, got %v", err)
	}
}
