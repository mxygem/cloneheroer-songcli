package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Filter handles filtering songs based on various criteria
type Filter struct {
	name    string
	artist  string
	genre   string
	charter string
	year    int
	length  string // e.g., ">5:00" or "<3:30"
	inst    string
}

// NewFilter creates a new Filter instance
func NewFilter(name, artist, genre, charter string, year int, length, inst string) *Filter {
	return &Filter{
		name:    name,
		artist:  artist,
		genre:   genre,
		charter: charter,
		year:    year,
		length:  length,
		inst:    inst,
	}
}

// Apply applies all filters to the song list
func (f *Filter) Apply(songs []*Song) []*Song {
	if f.isEmpty() {
		return songs
	}
	
	var filtered []*Song
	for _, song := range songs {
		if f.matches(song) {
			filtered = append(filtered, song)
		}
	}
	
	return filtered
}

// isEmpty checks if any filters are set
func (f *Filter) isEmpty() bool {
	return f.name == "" && f.artist == "" && f.genre == "" && 
		f.charter == "" && f.year == 0 && f.length == "" && f.inst == ""
}

// matches checks if a song matches all filter criteria
func (f *Filter) matches(song *Song) bool {
	if f.name != "" && !fuzzyMatch(song.Name, f.name) {
		return false
	}
	
	if f.artist != "" && !strings.Contains(strings.ToLower(song.Artist), strings.ToLower(f.artist)) {
		return false
	}
	
	if f.genre != "" && !strings.Contains(strings.ToLower(song.Genre), strings.ToLower(f.genre)) {
		return false
	}
	
	if f.charter != "" {
		charterMatch := false
		for _, charter := range song.Charters {
			if strings.Contains(strings.ToLower(charter), strings.ToLower(f.charter)) {
				charterMatch = true
				break
			}
		}
		if !charterMatch {
			return false
		}
	}
	
	if f.year != 0 && song.Year != f.year {
		return false
	}
	
	if f.length != "" && !f.matchesLength(song) {
		return false
	}
	
	if f.inst != "" && !f.matchesInstrument(song) {
		return false
	}
	
	return true
}

// matchesLength checks if song length matches the filter
func (f *Filter) matchesLength(song *Song) bool {
	if f.length == "" {
		return true
	}
	
	// Parse length filter (e.g., ">5:00", "<3:30", "=2:15")
	re := regexp.MustCompile(`^([><=]+)(\d+):(\d+)$`)
	matches := re.FindStringSubmatch(f.length)
	if len(matches) != 4 {
		return true // Invalid format, don't filter
	}
	
	op := matches[1]
	filterMinutes, _ := strconv.Atoi(matches[2])
	filterSeconds, _ := strconv.Atoi(matches[3])
	filterDuration := time.Duration(filterMinutes)*time.Minute + time.Duration(filterSeconds)*time.Second
	
	songDuration := song.Length
	
	switch op {
	case ">":
		return songDuration > filterDuration
	case ">=":
		return songDuration >= filterDuration
	case "<":
		return songDuration < filterDuration
	case "<=":
		return songDuration <= filterDuration
	case "=", "==":
		// Allow 5 second tolerance
		diff := songDuration - filterDuration
		if diff < 0 {
			diff = -diff
		}
		return diff < 5*time.Second
	default:
		return true
	}
}

// matchesInstrument checks if song has the specified instrument
func (f *Filter) matchesInstrument(song *Song) bool {
	if f.inst == "" {
		return true
	}
	
	inst := Instrument(strings.ToLower(f.inst))
	return song.HasInstrument(inst)
}

// fuzzyMatch performs simple fuzzy matching (substring match with case insensitivity)
// For better fuzzy matching, you could use a library like github.com/sahilm/fuzzy
func fuzzyMatch(text, pattern string) bool {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)
	
	// Simple substring match
	if strings.Contains(text, pattern) {
		return true
	}
	
	// Check if all pattern characters appear in order in text
	patternIdx := 0
	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			patternIdx++
		}
	}
	
	return patternIdx == len(pattern)
}

