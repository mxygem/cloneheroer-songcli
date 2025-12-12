package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Scanner handles scanning directories for songs and caching results
type Scanner struct {
	rootDir string
	cacheFile string
}

// CacheEntry represents a cached song entry
type CacheEntry struct {
	Path      string
	Name      string
	Artist    string
	Album     string
	Genre     string
	Year      int
	Charters  []string `json:"charters,omitempty"` // Multiple charters
	Charter   string   `json:"charter,omitempty"`   // Legacy single charter (for backward compatibility)
	Length    int64 // milliseconds
	Instruments map[string]int
	PreviewStart int64
	Icon       string
	LoadingPhrase string
	AlbumTrack int
	PlaylistTrack int
}

// Cache represents the cache file structure
type Cache struct {
	Hash  string
	Songs []CacheEntry
}

// NewScanner creates a new Scanner instance
func NewScanner(rootDir string) *Scanner {
	cacheDir := filepath.Join(os.TempDir(), "cloneheroer")
	os.MkdirAll(cacheDir, 0755)
	hash := sha256.Sum256([]byte(rootDir))
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("cache_%x.json", hash[:8]))
	
	return &Scanner{
		rootDir:   rootDir,
		cacheFile: cacheFile,
	}
}

// LoadSongs loads songs from directory, using cache if available and valid
func (s *Scanner) LoadSongs() ([]*Song, error) {
	// Calculate directory hash
	currentHash, err := s.calculateDirHash()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate directory hash: %w", err)
	}
	
	// Try to load from cache
	if cached, err := s.loadCache(); err == nil && cached.Hash == currentHash {
		return s.convertCacheToSongs(cached), nil
	}
	
	// Cache miss or invalid, scan directory
	songs, err := s.scanDirectory()
	if err != nil {
		return nil, err
	}
	
	// Save to cache
	if err := s.saveCache(currentHash, songs); err != nil {
		// Log but don't fail - caching is optional
		fmt.Fprintf(os.Stderr, "Warning: failed to save cache: %v\n", err)
	}
	
	return songs, nil
}

// calculateDirHash calculates a hash of the directory structure
func (s *Scanner) calculateDirHash() (string, error) {
	hash := sha256.New()
	
	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Include file paths and modification times in hash
		relPath, _ := filepath.Rel(s.rootDir, path)
		hash.Write([]byte(relPath))
		hash.Write([]byte(info.ModTime().String()))
		
		return nil
	})
	
	if err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// scanDirectory recursively scans for song.ini files
func (s *Scanner) scanDirectory() ([]*Song, error) {
	var songs []*Song
	
	err := filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		if strings.HasSuffix(strings.ToLower(path), "song.ini") {
			song, err := ParseSong(path)
			if err != nil {
				// Log but continue - some files might be malformed
				fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
				return nil
			}
			songs = append(songs, song)
		}
		
		return nil
	})
	
	return songs, err
}

// loadCache loads the cache from disk
func (s *Scanner) loadCache() (*Cache, error) {
	file, err := os.Open(s.cacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var cache Cache
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cache); err != nil {
		return nil, err
	}
	
	return &cache, nil
}

// saveCache saves songs to cache
func (s *Scanner) saveCache(hash string, songs []*Song) error {
	cache := Cache{
		Hash:  hash,
		Songs: make([]CacheEntry, len(songs)),
	}
	
	for i, song := range songs {
		instruments := make(map[string]int)
		for inst, diff := range song.Instruments {
			instruments[string(inst)] = diff
		}
		
		cache.Songs[i] = CacheEntry{
			Path:        song.Path,
			Name:        song.Name,
			Artist:      song.Artist,
			Album:       song.Album,
			Genre:       song.Genre,
			Year:        song.Year,
			Charters:    song.Charters,
			Length:      int64(song.Length / time.Millisecond),
			Instruments: instruments,
			PreviewStart: song.PreviewStart,
			Icon:        song.Icon,
			LoadingPhrase: song.LoadingPhrase,
			AlbumTrack:   song.AlbumTrack,
			PlaylistTrack: song.PlaylistTrack,
		}
	}
	
	file, err := os.Create(s.cacheFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cache)
}

// convertCacheToSongs converts cache entries back to Song structs
func (s *Scanner) convertCacheToSongs(cache *Cache) []*Song {
	songs := make([]*Song, len(cache.Songs))
	
	for i, entry := range cache.Songs {
		instruments := make(map[Instrument]int)
		for instStr, diff := range entry.Instruments {
			instruments[Instrument(instStr)] = diff
		}
		
		charters := entry.Charters
		if len(charters) == 0 && entry.Charter != "" {
			// Handle old cache format with single Charter field
			charters = []string{entry.Charter}
		}
		
		songs[i] = &Song{
			Path:        entry.Path,
			Name:        entry.Name,
			Artist:      entry.Artist,
			Album:       entry.Album,
			Genre:       entry.Genre,
			Year:        entry.Year,
			Charters:    charters,
			Length:      time.Duration(entry.Length) * time.Millisecond,
			Instruments: instruments,
			PreviewStart: entry.PreviewStart,
			Icon:        entry.Icon,
			LoadingPhrase: entry.LoadingPhrase,
			AlbumTrack:   entry.AlbumTrack,
			PlaylistTrack: entry.PlaylistTrack,
		}
	}
	
	return songs
}

