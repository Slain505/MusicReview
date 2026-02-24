package audioanalyzer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	DurationMS int
	Peaks      []int // 0..1000
}

func Analyze(ctx context.Context, path string, peaksCount int) (Result, error) {
	if peaksCount <= 0 {
		peaksCount = 1000
	}

	durMS, err := durationWithFFProbe(ctx, path)
	if err != nil {
		return Result{}, err
	}

	peaks, err := peaksWithFFMpeg(ctx, path, durMS, peaksCount)
	if err != nil {
		// duration good even if peaks unsuccessful
		return Result{DurationMS: durMS, Peaks: nil}, nil
	}

	return Result{DurationMS: durMS, Peaks: peaks}, nil
}

func durationWithFFProbe(ctx context.Context, path string) (int, error) {
	// ffprobe -show_entries format=duration returns seconds float
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0, errors.New("ffprobe duration empty")
	}
	sec, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return int(math.Round(sec * 1000)), nil
}

func peaksWithFFMpeg(ctx context.Context, path string, durationMS int, peaksCount int) ([]int, error) {
	// Decode in PCM: mono, 8kHz, signed 16-bit little endian
	// ffmpeg -i file -f s16le -ac 1 -ar 8000 pipe:1
	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error",
		"-i", path,
		"-f", "s16le",
		"-ac", "1",
		"-ar", "8000",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	defer func() { _ = stdout.Close() }()

	sampleRate := 8000
	totalSamples := int(math.Round(float64(durationMS) * float64(sampleRate) / 1000.0))
	if totalSamples <= 0 {
		_ = cmd.Wait()
		return nil, errors.New("invalid totalSamples")
	}

	window := totalSamples / peaksCount
	if window <= 0 {
		window = 1
	}

	peaks := make([]int, 0, peaksCount)
	maxInWindow := int16(0)
	countInWindow := 0

	reader := bufio.NewReaderSize(stdout, 64*1024)
	buf := make([]byte, 2)

	for {
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			_ = cmd.Wait()
			return nil, err
		}

		s := int16(binary.LittleEndian.Uint16(buf))
		abs := s
		if abs < 0 {
			abs = -abs
		}
		if abs > maxInWindow {
			maxInWindow = abs
		}

		countInWindow++
		if countInWindow >= window {
			// normalize int16 max (0..32767) -> 0..1000
			v := int(math.Round(float64(maxInWindow) * 1000.0 / 32767.0))
			if v < 0 {
				v = 0
			}
			if v > 1000 {
				v = 1000
			}
			peaks = append(peaks, v)
			maxInWindow = 0
			countInWindow = 0
			if len(peaks) >= peaksCount {
				break
			}
		}
	}

	// finish reading process
	if err := cmd.Wait(); err != nil {
		_ = stderr.String()
	}

	// finish array length to peaksCount
	for len(peaks) < peaksCount {
		peaks = append(peaks, 0)
	}
	return peaks, nil
}
