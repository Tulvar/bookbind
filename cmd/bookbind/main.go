package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Tulvar/bookbind/internal/app"
	"github.com/Tulvar/bookbind/pkg/version"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "bookbind:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	application := app.New()

	switch args[0] {
	case "inspect":
		return runInspect(ctx, application, args[1:], stdout)
	case "convert":
		return runConvert(ctx, application, args[1:], stdout)
	case "template":
		return runTemplate(ctx, application, args[1:], stdout)
	case "version":
		fmt.Fprintf(stdout, "bookbind %s\n", version.Version)
		return nil
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runTemplate(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("template", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	output := fs.String("output", "", "metadata yaml output path")
	overwrite := fs.Bool("overwrite", false, "overwrite output if it exists")

	if err := fs.Parse(reorderFlagArgs(args, map[string]bool{
		"output":    true,
		"overwrite": false,
	})); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("template expects exactly one input path")
	}

	result, err := application.TemplateMetadata(ctx, app.TemplateRequest{
		InputPath:  fs.Arg(0),
		OutputPath: *output,
		Overwrite:  *overwrite,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Input: %s\n", result.InputPath)
	fmt.Fprintf(stdout, "Metadata: %s\n", result.OutputPath)
	if result.Book.Title != "" {
		fmt.Fprintf(stdout, "Title: %s\n", result.Book.Title)
	}
	if result.Book.Cover != "" {
		fmt.Fprintf(stdout, "Cover: %s\n", result.Book.Cover)
	}
	fmt.Fprintln(stdout, "Status: written")
	return nil
}

func runInspect(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("inspect expects exactly one input path")
	}

	result, err := application.InspectInput(ctx, app.InspectRequest{InputPath: fs.Arg(0)})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Input: %s\n", result.Input.Path)
	fmt.Fprintf(stdout, "Files: %d\n", len(result.Input.Files))
	for _, file := range result.Input.Files {
		fmt.Fprintf(stdout, "  - %s", file.Path)
		if file.Duration > 0 {
			fmt.Fprintf(stdout, " (%s", formatDuration(file.Duration))
			if file.Codec != "" {
				fmt.Fprintf(stdout, ", %s", file.Codec)
			}
			if file.Bitrate > 0 {
				fmt.Fprintf(stdout, ", %d kbps", file.Bitrate/1000)
			}
			if file.Channels > 0 {
				fmt.Fprintf(stdout, ", %d ch", file.Channels)
			}
			fmt.Fprint(stdout, ")")
		}
		fmt.Fprintln(stdout)
	}
	return nil
}

func runConvert(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	output := fs.String("output", "", "output m4b path")
	metadata := fs.String("metadata", "", "metadata yaml path")
	cover := fs.String("cover", "", "cover image path")
	dryRun := fs.Bool("dry-run", false, "print planned work without creating output")
	overwrite := fs.Bool("overwrite", false, "overwrite output if it exists")

	if err := fs.Parse(reorderFlagArgs(args, map[string]bool{
		"output":    true,
		"metadata":  true,
		"cover":     true,
		"dry-run":   false,
		"overwrite": false,
	})); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("convert expects exactly one input path")
	}

	result, err := application.Convert(ctx, app.ConvertRequest{
		InputPath:    fs.Arg(0),
		OutputPath:   *output,
		MetadataPath: *metadata,
		CoverPath:    *cover,
		DryRun:       *dryRun,
		Overwrite:    *overwrite,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Input: %s\n", result.Input.Path)
	fmt.Fprintf(stdout, "Files: %d\n", len(result.Input.Files))
	if result.MetadataPath != "" {
		fmt.Fprintf(stdout, "Metadata: %s\n", result.MetadataPath)
	}
	if result.Metadata.Title != "" {
		fmt.Fprintf(stdout, "Title: %s\n", result.Metadata.Title)
	}
	if result.CoverPath != "" {
		fmt.Fprintf(stdout, "Cover: %s\n", result.CoverPath)
	}
	fmt.Fprintf(stdout, "Output: %s\n", result.OutputPath)
	if result.DryRun {
		fmt.Fprintln(stdout, "Mode: dry-run")
	}
	if len(result.Command) > 0 {
		fmt.Fprintf(stdout, "Command: %s\n", strings.Join(result.Command, " "))
	}
	if result.DryRun {
		fmt.Fprintln(stdout, "Status: planned")
	} else {
		fmt.Fprintln(stdout, "Status: done")
	}
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "bookbind converts MP3 audiobooks to M4B.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  bookbind inspect <mp3-or-directory>")
	fmt.Fprintln(w, "  bookbind convert <mp3-or-directory> [--output book.m4b] [--dry-run]")
	fmt.Fprintln(w, "  bookbind template <mp3-or-directory> [--output bookbind.yaml]")
	fmt.Fprintln(w, "  bookbind version")
}

func formatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm%02ds", minutes, seconds)
}

func reorderFlagArgs(args []string, valueFlags map[string]bool) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}

		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		if strings.Contains(arg, "=") || !valueFlags[name] {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}

	return append(flags, positionals...)
}
