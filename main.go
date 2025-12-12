package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "cloneheroer",
		Short: "Lookup song charts for Clone Hero",
		Long:  "A CLI tool to search and filter Clone Hero song charts. If no directory is specified, uses the current directory.",
		Args:  cobra.NoArgs,
		RunE:  run,
	}

	// Flags
	directory     string
	outputFile    string
	countOnly     bool
	filterName    string
	filterArtist  string
	filterGenre   string
	filterCharter string
	filterYear    int
	filterLength  string
	filterInst    string
	sortBy        string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&directory, "directory", "d", ".", "Directory to recursively search for songs (default: current directory)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Write results to file instead of stdout")
	rootCmd.PersistentFlags().BoolVarP(&countOnly, "count", "c", false, "Only return count of matching songs")
	rootCmd.PersistentFlags().StringVarP(&filterName, "name", "n", "", "Filter by song name (fuzzy matching)")
	rootCmd.PersistentFlags().StringVarP(&filterArtist, "artist", "a", "", "Filter by artist")
	rootCmd.PersistentFlags().StringVarP(&filterGenre, "genre", "g", "", "Filter by genre")
	rootCmd.PersistentFlags().StringVarP(&filterCharter, "charter", "", "", "Filter by charter")
	rootCmd.PersistentFlags().IntVarP(&filterYear, "year", "y", 0, "Filter by year")
	rootCmd.PersistentFlags().StringVarP(&filterLength, "length", "l", "", "Filter by song length (e.g., '>5:00' or '<3:30')")
	rootCmd.PersistentFlags().StringVarP(&filterInst, "instrument", "i", "", "Filter by instrument (guitar, drums, bass, etc.)")
	rootCmd.PersistentFlags().StringVarP(&sortBy, "sort", "s", "", "Sort by field (name, artist, year, length, genre, charter)")
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize scanner
	scanner := NewScanner(directory)

	// Load songs (with caching)
	songs, err := scanner.LoadSongs()
	if err != nil {
		return fmt.Errorf("failed to load songs: %w", err)
	}

	// Apply filters
	filter := NewFilter(filterName, filterArtist, filterGenre, filterCharter, filterYear, filterLength, filterInst)
	filteredSongs := filter.Apply(songs)

	// Sort
	if sortBy != "" {
		sorter := NewSorter(sortBy)
		sorter.Sort(filteredSongs)
	}

	// Output
	output := NewOutput(outputFile, countOnly)
	return output.Write(songs, filteredSongs)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
