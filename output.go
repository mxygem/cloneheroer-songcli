package main

import (
	"fmt"
	"html"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// Output handles writing results to stdout or file
type Output struct {
	writer    io.Writer
	countOnly bool
}

// NewOutput creates a new Output instance
func NewOutput(outputFile string, countOnly bool) *Output {
	var writer io.Writer = os.Stdout

	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create output file: %v\n", err)
		} else {
			writer = file
		}
	}

	return &Output{
		writer:    writer,
		countOnly: countOnly,
	}
}

// Write writes the results
func (o *Output) Write(allSongs, filteredSongs []*Song) error {
	if o.countOnly {
		fmt.Fprintf(o.writer, "%d\n", len(filteredSongs))
		return nil
	}

	// Write summary
	fmt.Fprintf(o.writer, "Found %d song(s) (out of %d total)\n\n", len(filteredSongs), len(allSongs))

	if len(filteredSongs) == 0 {
		return nil
	}

	// Write songs
	for i, song := range filteredSongs {
		o.writeSong(song, i+1)
		fmt.Fprintln(o.writer)
	}

	return nil
}

// writeSong writes a single song entry
func (o *Output) writeSong(song *Song, index int) {
	fmt.Fprintf(o.writer, "%d. %s\n", index, color.New(color.Bold).Sprint(song.Name))
	fmt.Fprintf(o.writer, "   Artist: %s\n", song.Artist)
	if song.Album != "" {
		fmt.Fprintf(o.writer, "   Album: %s\n", song.Album)
	}
	if song.Genre != "" {
		fmt.Fprintf(o.writer, "   Genre: %s\n", song.Genre)
	}
	if song.Year > 0 {
		fmt.Fprintf(o.writer, "   Year: %d\n", song.Year)
	}
	if len(song.Charters) > 0 {
		// Format all charters with colors
		formattedCharters := make([]string, len(song.Charters))
		for i, charter := range song.Charters {
			formattedCharters[i] = o.formatCharter(charter)
		}
		charterStr := strings.Join(formattedCharters, ", ")
		fmt.Fprintf(o.writer, "   Charter: %s\n", charterStr)
	}
	fmt.Fprintf(o.writer, "   Length: %s\n", song.FormatLength())

	instruments := song.InstrumentList()
	if instruments != "" {
		fmt.Fprintf(o.writer, "   Instruments: %s\n", instruments)
	}

	// Show path relative to current directory
	wd, _ := os.Getwd()
	if strings.HasPrefix(song.Path, wd) {
		relPath := strings.TrimPrefix(song.Path, wd+"/")
		fmt.Fprintf(o.writer, "   Path: %s\n", relPath)
	} else {
		fmt.Fprintf(o.writer, "   Path: %s\n", song.Path)
	}
}

// formatCharter formats charter name, handling HTML colors
func (o *Output) formatCharter(charter string) string {
	if charter == "" {
		return ""
	}
	
	// Check for HTML color tags (e.g., <color=#FF0000>text</color>)
	// Use a more robust regex that handles multiple consecutive tags
	re := regexp.MustCompile(`<color=#([0-9A-Fa-f]{6})>(.*?)</color>`)

	// If we're writing to a file or stdout without color support, strip HTML
	if o.writer != os.Stdout {
		// Remove HTML tags
		charter = re.ReplaceAllString(charter, "$2")
		charter = html.UnescapeString(charter)
		return charter
	}

	// Replace all HTML color tags with ANSI colors
	// Find all matches and replace them
	var result strings.Builder
	lastIndex := 0
	
	for _, match := range re.FindAllStringSubmatchIndex(charter, -1) {
		// Add text before the match
		result.WriteString(charter[lastIndex:match[0]])
		
		// Extract color hex and text
		colorHex := charter[match[2]:match[3]]
		text := charter[match[4]:match[5]]
		
		// Convert hex to RGB
		var r, g, b uint8
		if _, err := fmt.Sscanf(colorHex, "%02x%02x%02x", &r, &g, &b); err == nil {
			// Create color function and write colored text
			c := color.RGB(int(r), int(g), int(b))
			result.WriteString(c.Sprint(text))
		} else {
			// Fallback to plain text if color parsing fails
			result.WriteString(text)
		}
		
		lastIndex = match[1]
	}
	
	// Add remaining text after last match
	result.WriteString(charter[lastIndex:])
	
	// Handle any remaining HTML entities
	finalResult := html.UnescapeString(result.String())
	
	return finalResult
}
