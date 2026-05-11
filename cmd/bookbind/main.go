package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Tulvar/bookbind/internal/app"
	"github.com/Tulvar/bookbind/internal/audio"
	"github.com/Tulvar/bookbind/internal/metadata"
	"github.com/Tulvar/bookbind/internal/providers"
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
	case "search":
		return runSearch(ctx, application, args[1:], stdout)
	case "providers":
		return runProviders(stdout)
	case "metadata":
		return runMetadata(ctx, application, args[1:], stdout)
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

func runSearch(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	title := fs.String("title", "", "book title")
	author := fs.String("author", "", "book author")
	provider := fs.String("provider", "", "comma-separated metadata providers")

	if err := fs.Parse(reorderFlagArgs(args, map[string]bool{
		"title":    true,
		"author":   true,
		"provider": true,
	})); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("search does not accept positional arguments")
	}

	result, err := application.SearchMetadata(ctx, app.SearchRequest{
		Title:     *title,
		Author:    *author,
		Providers: splitProviderList(*provider),
	})
	if err != nil {
		return err
	}

	return printSearchCandidates(stdout, result.Candidates)
}

func printSearchCandidates(stdout io.Writer, candidates []providers.Candidate) error {
	fmt.Fprintf(stdout, "Candidates: %d\n", len(candidates))
	if len(candidates) == 0 {
		return nil
	}

	table := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "#\tProvider\tID\tTitle\tAuthors\tYear\tConfidence")
	for i, candidate := range candidates {
		fmt.Fprintf(
			table,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			i+1,
			candidate.Provider,
			candidate.ID,
			candidate.Title,
			strings.Join(candidate.Authors, ", "),
			formatYear(candidate.Year),
			formatConfidence(candidate.Confidence),
		)
	}
	return table.Flush()
}

func runMetadata(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("metadata", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	provider := fs.String("provider", "", "metadata provider")
	id := fs.String("id", "", "provider candidate id")
	output := fs.String("output", "bookbind.yaml", "metadata yaml output path")
	preview := fs.Bool("preview", false, "print candidate details without writing metadata")
	overwrite := fs.Bool("overwrite", false, "overwrite output if it exists")

	if err := fs.Parse(reorderFlagArgs(args, map[string]bool{
		"provider":  true,
		"id":        true,
		"output":    true,
		"preview":   false,
		"overwrite": false,
	})); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("metadata does not accept positional arguments")
	}

	if *preview {
		result, err := application.PreviewMetadata(ctx, app.PreviewMetadataRequest{
			Provider: *provider,
			ID:       *id,
		})
		if err != nil {
			return err
		}
		printMetadataDetails(stdout, result.Candidate, result.Book)
		return nil
	}

	result, err := application.ResolveMetadata(ctx, app.ResolveMetadataRequest{
		Provider:   *provider,
		ID:         *id,
		OutputPath: *output,
		Overwrite:  *overwrite,
	})
	if err != nil {
		return err
	}

	printMetadataDetails(stdout, result.Candidate, result.Book)
	fmt.Fprintf(stdout, "Metadata: %s\n", result.OutputPath)
	fmt.Fprintln(stdout, "Status: written")
	return nil
}

func printMetadataDetails(stdout io.Writer, candidate providers.Candidate, book metadata.Book) {
	fmt.Fprintf(stdout, "Provider: %s\n", candidate.Provider)
	fmt.Fprintf(stdout, "ID: %s\n", candidate.ID)
	if book.Title != "" {
		fmt.Fprintf(stdout, "Title: %s\n", book.Title)
	}
	if len(book.NormalizedAuthors()) > 0 {
		fmt.Fprintf(stdout, "Author: %s\n", strings.Join(book.NormalizedAuthors(), ", "))
	}
	if len(book.NormalizedNarrators()) > 0 {
		fmt.Fprintf(stdout, "Narrator: %s\n", strings.Join(book.NormalizedNarrators(), ", "))
	}
	if book.Series != "" {
		fmt.Fprintf(stdout, "Series: %s", book.Series)
		if book.SeriesIndex != "" {
			fmt.Fprintf(stdout, " #%s", book.SeriesIndex)
		}
		fmt.Fprintln(stdout)
	}
	if book.PublishedYear > 0 {
		fmt.Fprintf(stdout, "Published year: %d\n", book.PublishedYear)
	}
	if book.Cover != "" {
		fmt.Fprintf(stdout, "Cover: %s\n", book.Cover)
	}
	if candidate.Confidence > 0 {
		fmt.Fprintf(stdout, "Confidence: %.2f\n", candidate.Confidence)
	}
}

func runProviders(stdout io.Writer) error {
	for _, provider := range app.AvailableProviders() {
		status := "disabled"
		if provider.Enabled {
			status = "enabled"
		}
		fmt.Fprintf(stdout, "%s\t%s\n", provider.Name, status)
	}
	return nil
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
		printEmbeddedTags(stdout, file.Tags)
		printChapters(stdout, file.Chapters)
	}
	return nil
}

func printEmbeddedTags(stdout io.Writer, tags audio.EmbeddedTags) {
	if tags.Title == "" &&
		tags.Artist == "" &&
		tags.Album == "" &&
		tags.Composer == "" &&
		tags.Genre == "" &&
		tags.Date == "" &&
		tags.Comment == "" &&
		tags.Language == "" {
		return
	}

	fmt.Fprintln(stdout, "    Embedded metadata:")
	printTag(stdout, "title", tags.Title)
	printTag(stdout, "artist", tags.Artist)
	printTag(stdout, "album", tags.Album)
	printTag(stdout, "composer", tags.Composer)
	printTag(stdout, "genre", tags.Genre)
	printTag(stdout, "date", tags.Date)
	printTag(stdout, "language", tags.Language)
	printTag(stdout, "comment", tags.Comment)
}

func printTag(stdout io.Writer, name, value string) {
	if value != "" {
		fmt.Fprintf(stdout, "      %s: %s\n", name, value)
	}
}

func printChapters(stdout io.Writer, chapters []audio.Chapter) {
	if len(chapters) == 0 {
		return
	}

	fmt.Fprintf(stdout, "    Chapters: %d\n", len(chapters))
	for _, chapter := range chapters {
		fmt.Fprintf(stdout, "      - %s", chapter.Title)
		if chapter.End > 0 {
			fmt.Fprintf(stdout, " (%s - %s)", formatTimecode(chapter.Start), formatTimecode(chapter.End))
		}
		fmt.Fprintln(stdout)
	}
}

func runConvert(ctx context.Context, application *app.App, args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	output := fs.String("output", "", "output m4b path")
	metadata := fs.String("metadata", "", "metadata yaml path")
	cover := fs.String("cover", "", "cover image path")
	chapterEvery := fs.String("chapter-every", "", "create synthetic chapters at the given interval, for example 10m")
	dryRun := fs.Bool("dry-run", false, "print planned work without creating output")
	overwrite := fs.Bool("overwrite", false, "overwrite output if it exists")

	if err := fs.Parse(reorderFlagArgs(args, map[string]bool{
		"output":        true,
		"metadata":      true,
		"cover":         true,
		"chapter-every": true,
		"dry-run":       false,
		"overwrite":     false,
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
		ChapterEvery: *chapterEvery,
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
	if result.ChapterEvery != "" {
		fmt.Fprintf(stdout, "Chapter every: %s\n", result.ChapterEvery)
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
	fmt.Fprintln(w, "  bookbind providers")
	fmt.Fprintln(w, "  bookbind search --title <title> [--author <author>] [--provider openlibrary,googlebooks]")
	fmt.Fprintln(w, "  bookbind metadata --provider <provider> --id <candidate-id> [--preview] [--output bookbind.yaml]")
	fmt.Fprintln(w, "  bookbind convert <mp3-or-directory> [--output book.m4b] [--dry-run] [--chapter-every 10m]")
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

func formatTimecode(duration time.Duration) string {
	duration = duration.Round(time.Millisecond)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	millis := int(duration.Milliseconds()) % 1000
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, millis)
	}
	return fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, millis)
}

func formatYear(year int) string {
	if year <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", year)
}

func formatConfidence(confidence float64) string {
	if confidence <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", confidence)
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

func splitProviderList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	providers := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			providers = append(providers, name)
		}
	}
	return providers
}
