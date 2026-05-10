package audio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Inspector struct {
	prober Prober
}

type InspectorOption func(*Inspector)

func NewInspector(options ...InspectorOption) *Inspector {
	inspector := &Inspector{
		prober: NewFFProbe("ffprobe"),
	}
	for _, option := range options {
		option(inspector)
	}
	return inspector
}

func WithProber(prober Prober) InspectorOption {
	return func(i *Inspector) {
		i.prober = prober
	}
}

func (i *Inspector) Inspect(ctx context.Context, inputPath string) (Input, error) {
	if strings.TrimSpace(inputPath) == "" {
		return Input{}, fmt.Errorf("input path is required")
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		return Input{}, err
	}

	var files []File
	if info.IsDir() {
		files, err = i.inspectDirectory(ctx, inputPath)
	} else {
		files, err = i.inspectFile(ctx, inputPath)
	}
	if err != nil {
		return Input{}, err
	}
	if len(files) == 0 {
		return Input{}, fmt.Errorf("no mp3 files found in %s", inputPath)
	}

	return Input{
		Path:  inputPath,
		Files: files,
	}, nil
}

func (i *Inspector) inspectDirectory(ctx context.Context, dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if !isMP3(path) {
			continue
		}
		file, err := i.inspectFile(ctx, path)
		if err != nil {
			return nil, err
		}
		files = append(files, file...)
	}

	sort.SliceStable(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

func (i *Inspector) inspectFile(ctx context.Context, path string) ([]File, error) {
	if !isMP3(path) {
		return nil, fmt.Errorf("input file must be an mp3: %s", path)
	}

	probe, err := i.probe(ctx, path)
	if err != nil {
		return nil, err
	}

	return []File{{
		Path:     path,
		Name:     filepath.Base(path),
		Duration: probe.Duration,
		Codec:    probe.Codec,
		Bitrate:  probe.Bitrate,
		Channels: probe.Channels,
		Tags:     probe.Tags,
	}}, nil
}

func (i *Inspector) probe(ctx context.Context, path string) (ProbeResult, error) {
	if i.prober == nil {
		return ProbeResult{}, nil
	}
	return i.prober.Probe(ctx, path)
}

func isMP3(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".mp3")
}
