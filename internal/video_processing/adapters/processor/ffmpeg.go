package processor

import (
	"context"
	"fmt"
	"os/exec"
)

type FFmpeg struct {
	Binary string
}

func NewFFmpeg(binary string) *FFmpeg {
	return &FFmpeg{Binary: binary}
}

func (f *FFmpeg) Process(ctx context.Context, input string, output string) error {
	command := exec.CommandContext(ctx, f.Binary, "-y", "-i", input, "-c:v", "libx264", "-c:a", "aac", output)
	if outputBytes, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, outputBytes)
	}
	return nil
}
