package processor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

type FFmpeg struct {
	Binary          string
	FramesPerSecond string
}

func NewFFmpeg(binary string, framesPerSecond string) *FFmpeg {
	return &FFmpeg{Binary: binary, FramesPerSecond: framesPerSecond}
}

func (f *FFmpeg) Process(ctx context.Context, input string, outputDir string) ([]string, error) {
	outputPattern := filepath.Join(outputDir, "frame_%06d.jpg")
	command := exec.CommandContext(ctx, f.Binary, "-y", "-i", input, "-vf", "fps="+f.FramesPerSecond, "-q:v", "2", outputPattern)
	if outputBytes, err := command.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w: %s", err, outputBytes)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "frame_*.jpg"))
	if err != nil {
		return nil, fmt.Errorf("listar frames gerados: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("ffmpeg não gerou imagens")
	}
	return files, nil
}
