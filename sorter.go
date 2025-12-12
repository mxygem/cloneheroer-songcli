package main

import (
	"sort"
	"strings"
)

// Sorter handles sorting songs by various fields
type Sorter struct {
	sortBy string
}

// NewSorter creates a new Sorter instance
func NewSorter(sortBy string) *Sorter {
	return &Sorter{sortBy: strings.ToLower(sortBy)}
}

// Sort sorts the songs slice in place
func (s *Sorter) Sort(songs []*Song) {
	sort.Slice(songs, func(i, j int) bool {
		return s.less(songs[i], songs[j])
	})
}

// less compares two songs based on the sort field
func (s *Sorter) less(a, b *Song) bool {
	switch s.sortBy {
	case "name":
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "artist":
		if a.Artist != b.Artist {
			return strings.ToLower(a.Artist) < strings.ToLower(b.Artist)
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "year":
		if a.Year != b.Year {
			return a.Year < b.Year
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "length":
		if a.Length != b.Length {
			return a.Length < b.Length
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "genre":
		if a.Genre != b.Genre {
			return strings.ToLower(a.Genre) < strings.ToLower(b.Genre)
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case "charter":
		charterA := ""
		charterB := ""
		if len(a.Charters) > 0 {
			charterA = a.Charters[0]
		}
		if len(b.Charters) > 0 {
			charterB = b.Charters[0]
		}
		if charterA != charterB {
			return strings.ToLower(charterA) < strings.ToLower(charterB)
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	default:
		// Default: sort by artist, then name
		if a.Artist != b.Artist {
			return strings.ToLower(a.Artist) < strings.ToLower(b.Artist)
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}
}

