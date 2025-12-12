package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

// Instrument represents available instruments in Clone Hero
type Instrument string

const (
	InstrumentGuitar    Instrument = "guitar"
	InstrumentRhythm    Instrument = "rhythm"
	InstrumentBass      Instrument = "bass"
	InstrumentDrums     Instrument = "drums"
	InstrumentKeys      Instrument = "keys"
	InstrumentBand      Instrument = "band"
	InstrumentGuitarGHL Instrument = "guitarghl"
	InstrumentBassGHL   Instrument = "bassghl"
)

// Song represents a Clone Hero song chart
type Song struct {
	Path          string
	Name          string
	Artist        string
	Album         string
	Genre         string
	Year          int
	Charters      []string           // Multiple charters (comma-separated in ini file)
	Length        time.Duration      // in milliseconds, converted to Duration
	Instruments   map[Instrument]int // instrument -> difficulty level
	PreviewStart  int64
	Icon          string
	LoadingPhrase string
	AlbumTrack    int
	PlaylistTrack int
}

// ParseSong parses a song.ini file and returns a Song struct
func ParseSong(path string) (*Song, error) {
	// First, try to load with ini library
	cfg, err := ini.Load(path)
	if err != nil {
		// If loading fails, try manual parsing for malformed files
		return parseSongManually(path)
	}

	// Try case-insensitive section lookup
	section := cfg.Section("song")
	if section == nil {
		section = cfg.Section("Song")
	}
	if section == nil {
		return nil, fmt.Errorf("no [song] section found")
	}

	song := &Song{
		Path:        path,
		Instruments: make(map[Instrument]int),
		Charters:    []string{},
	}

	// Parse basic fields
	song.Name = section.Key("name").String()
	song.Artist = section.Key("artist").String()
	song.Album = section.Key("album").String()
	song.Genre = section.Key("genre").String()

	// Parse charter(s) - can be comma or ampersand separated
	// The INI library has issues with HTML tags, so ALWAYS read directly from file
	charterStr := readCharterDirectly(path)
	if charterStr == "" {
		// Fallback to manual function if direct read fails
		charterStr = getCharterValueManually(section, path)
	}

	if charterStr != "" {
		// Split by comma or ampersand (but not inside HTML tags)
		charters := splitCharters(charterStr)
		for _, c := range charters {
			c = strings.TrimSpace(c)
			if c != "" {
				song.Charters = append(song.Charters, c)
			}
		}
	}

	song.Icon = section.Key("icon").String()
	song.LoadingPhrase = section.Key("loading_phrase").String()

	// Parse year
	if yearStr := section.Key("year").String(); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			song.Year = year
		}
	}

	// Parse song length (in milliseconds)
	if lengthStr := section.Key("song_length").String(); lengthStr != "" {
		if lengthMs, err := strconv.ParseInt(lengthStr, 10, 64); err == nil {
			song.Length = time.Duration(lengthMs) * time.Millisecond
		}
	}

	// Parse preview start time
	if previewStr := section.Key("preview_start_time").String(); previewStr != "" {
		if preview, err := strconv.ParseInt(previewStr, 10, 64); err == nil {
			song.PreviewStart = preview
		}
	}

	// Also check for malformed preview_start_time (missing =)
	if previewStr := getMalformedKey(section, "preview_start_time"); previewStr != "" {
		if preview, err := strconv.ParseInt(previewStr, 10, 64); err == nil {
			song.PreviewStart = preview
		}
	}

	// Parse album/playlist track
	if trackStr := section.Key("album_track").String(); trackStr != "" {
		if track, err := strconv.Atoi(trackStr); err == nil {
			song.AlbumTrack = track
		}
	}
	if trackStr := section.Key("playlist_track").String(); trackStr != "" {
		if track, err := strconv.Atoi(trackStr); err == nil {
			song.PlaylistTrack = track
		}
	}

	// Parse instrument difficulties
	instrumentKeys := map[string]Instrument{
		"diff_guitar":    InstrumentGuitar,
		"diff_rhythm":    InstrumentRhythm,
		"diff_bass":      InstrumentBass,
		"diff_drums":     InstrumentDrums,
		"diff_keys":      InstrumentKeys,
		"diff_band":      InstrumentBand,
		"diff_guitarghl": InstrumentGuitarGHL,
		"diff_bassghl":   InstrumentBassGHL,
	}

	for key, inst := range instrumentKeys {
		if diffStr := section.Key(key).String(); diffStr != "" {
			if diff, err := strconv.Atoi(diffStr); err == nil && diff > 0 {
				song.Instruments[inst] = diff
			}
		}
	}

	return song, nil
}

// parseSongManually handles malformed INI files that the ini library can't parse
func parseSongManually(path string) (*Song, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	song := &Song{
		Path:        path,
		Instruments: make(map[Instrument]int),
		Charters:    []string{},
	}

	lines := strings.Split(string(data), "\n")
	inSongSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header (case-insensitive)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.ToLower(strings.Trim(line[1:len(line)-1], " "))
			inSongSection = (sectionName == "song")
			continue
		}

		if !inSongSection {
			continue
		}

		// Handle malformed entries (missing =)
		if strings.HasPrefix(line, "preview_start_time") && !strings.Contains(line, "=") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if preview, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					song.PreviewStart = preview
				}
			}
			continue
		}

		// Parse normal key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "name":
			song.Name = value
		case "artist":
			song.Artist = value
		case "album":
			song.Album = value
		case "genre":
			song.Genre = value
		case "charter":
			charters := splitCharters(value)
			for _, c := range charters {
				c = strings.TrimSpace(c)
				if c != "" {
					song.Charters = append(song.Charters, c)
				}
			}
		case "icon":
			song.Icon = value
		case "loading_phrase":
			song.LoadingPhrase = value
		case "year":
			if year, err := strconv.Atoi(value); err == nil {
				song.Year = year
			}
		case "song_length":
			if lengthMs, err := strconv.ParseInt(value, 10, 64); err == nil {
				song.Length = time.Duration(lengthMs) * time.Millisecond
			}
		case "preview_start_time":
			if preview, err := strconv.ParseInt(value, 10, 64); err == nil {
				song.PreviewStart = preview
			}
		case "album_track":
			if track, err := strconv.Atoi(value); err == nil {
				song.AlbumTrack = track
			}
		case "playlist_track":
			if track, err := strconv.Atoi(value); err == nil {
				song.PlaylistTrack = track
			}
		case "diff_guitar":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentGuitar] = diff
			}
		case "diff_rhythm":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentRhythm] = diff
			}
		case "diff_bass":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentBass] = diff
			}
		case "diff_drums":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentDrums] = diff
			}
		case "diff_keys":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentKeys] = diff
			}
		case "diff_band":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentBand] = diff
			}
		case "diff_guitarghl":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentGuitarGHL] = diff
			}
		case "diff_bassghl":
			if diff, err := strconv.Atoi(value); err == nil && diff > 0 {
				song.Instruments[InstrumentBassGHL] = diff
			}
		}
	}

	return song, nil
}

// getMalformedKey tries to find a key-value pair that might be malformed (missing =)
func getMalformedKey(section *ini.Section, keyName string) string {
	// This is a fallback - the ini library should handle most cases
	// But we check section keys directly for malformed entries
	for _, key := range section.Keys() {
		if strings.HasPrefix(key.Name(), keyName+" ") {
			// Found a key that starts with the name but has a space instead of =
			parts := strings.Fields(key.Name())
			if len(parts) >= 2 && parts[0] == keyName {
				return parts[1]
			}
		}
	}
	return ""
}

// HasInstrument checks if the song has a specific instrument chart
func (s *Song) HasInstrument(inst Instrument) bool {
	_, ok := s.Instruments[inst]
	return ok
}

// FormatLength formats the song length as hh:mm:ss
func (s *Song) FormatLength() string {
	totalSeconds := int(s.Length.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// InstrumentList returns a comma-separated list of available instruments
func (s *Song) InstrumentList() string {
	var instruments []string
	for inst, diff := range s.Instruments {
		if diff > 0 {
			instruments = append(instruments, string(inst))
		}
	}
	return strings.Join(instruments, ", ")
}

// getCharterValueManually reads the charter value directly from the file to avoid INI parsing issues with HTML
func getCharterValueManually(section *ini.Section, filePath string) string {
	// Always read manually since INI library has issues with HTML tags in values
	// Use absolute path to avoid issues
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		// Fallback to INI library if file read fails
		return section.Key("charter").String()
	}

	// Use simple line-by-line parsing since regex lookahead isn't supported
	content := string(data)
	lines := strings.Split(content, "\n")
	inSongSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for section header
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.ToLower(strings.Trim(trimmed[1:len(trimmed)-1], " "))
			inSongSection = (sectionName == "song")
			continue
		}

		if !inSongSection {
			continue
		}

		// Check if this is the charter line (case-insensitive)
		lowerLine := strings.ToLower(trimmed)
		if strings.HasPrefix(lowerLine, "charter") {
			// Find the = sign
			idx := strings.Index(line, "=")
			if idx >= 0 && idx < len(line)-1 {
				value := strings.TrimSpace(line[idx+1:])
				// Remove trailing comments
				if commentIdx := strings.Index(value, ";"); commentIdx >= 0 {
					value = strings.TrimSpace(value[:commentIdx])
				}
				return value
			}
		}

		// If we hit another section, stop
		if strings.HasPrefix(trimmed, "[") {
			break
		}
	}

	// Fallback to INI library
	return section.Key("charter").String()
}

// readCharterDirectly reads charter value directly from file as last resort
func readCharterDirectly(filePath string) string {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}

	// Handle both Unix and Windows line endings
	content := string(data)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	inSongSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.ToLower(strings.Trim(trimmed[1:len(trimmed)-1], " "))
			inSongSection = (sectionName == "song")
			continue
		}

		if !inSongSection {
			continue
		}

		// Check for charter line - be more flexible with matching
		lowerTrimmed := strings.ToLower(trimmed)
		if strings.Contains(lowerTrimmed, "charter") && strings.Contains(line, "=") {
			// Find the = sign in the original line
			idx := strings.Index(line, "=")
			if idx >= 0 && idx < len(line)-1 {
				// Get everything after the =
				value := line[idx+1:]
				// Trim whitespace
				value = strings.TrimSpace(value)
				// Remove trailing comments
				if commentIdx := strings.Index(value, ";"); commentIdx >= 0 {
					value = strings.TrimSpace(value[:commentIdx])
				}
				// Remove trailing newline/carriage return if present
				value = strings.TrimRight(value, "\r\n")
				return value
			}
		}

		// If we hit another section, stop looking
		if strings.HasPrefix(trimmed, "[") && trimmed != "" {
			break
		}
	}

	return ""
}

// splitCharters splits charter string by comma or ampersand, but not inside HTML tags
func splitCharters(charterStr string) []string {
	var result []string
	var current strings.Builder
	inTag := false

	for i, r := range charterStr {
		char := string(r)

		if char == "<" {
			inTag = true
			current.WriteRune(r)
		} else if char == ">" {
			inTag = false
			current.WriteRune(r)
		} else if !inTag && (char == "," || char == "&") {
			// Check if this is a separator (not part of HTML entity like &amp;)
			// Look ahead to see if it's followed by more characters (likely a separator)
			// or if it's part of an HTML entity
			if char == "&" {
				// Check if it's an HTML entity (like &amp; &lt; etc)
				remaining := charterStr[i:]
				if strings.HasPrefix(remaining, "&amp;") ||
					strings.HasPrefix(remaining, "&lt;") ||
					strings.HasPrefix(remaining, "&gt;") ||
					strings.HasPrefix(remaining, "&quot;") ||
					strings.HasPrefix(remaining, "&#") {
					current.WriteRune(r)
					continue
				}
			}

			// It's a separator
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
