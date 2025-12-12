# Clone Hero Song CLI

A command-line tool to search and filter Clone Hero song charts.

## Installation

```bash
go build -o cloneheroer
```

To use the `ch` shorthand, create a symlink:
```bash
ln -s cloneheroer ch
```

## Usage

```bash
cloneheroer [directory] [flags]
```

Or with the shorthand:
```bash
cloneheroer ch [directory] [flags]
```

## Features

- **Caching**: Automatically caches song metadata to disk for faster subsequent runs
- **Hash-based invalidation**: Only rescans directories when files have changed
- **Filtering**: Filter songs by:
  - Song name (fuzzy matching)
  - Artist
  - Genre
  - Charter
  - Year
  - Song length (e.g., `>5:00`, `<3:30`)
  - Instrument (guitar, drums, bass, rhythm, keys, band, guitarghl, bassghl)
- **Sorting**: Sort results by name, artist, year, length, genre, or charter
- **Colored output**: Charter names with HTML color tags are converted to ANSI colors
- **Count mode**: Get just the count of matching songs
- **File output**: Write results to a file instead of stdout

## Examples

List all songs:
```bash
cloneheroer ./songs
```

Filter by artist:
```bash
cloneheroer ./songs --artist "Polyphia"
```

Filter by instrument:
```bash
cloneheroer ./songs --instrument drums
```

Filter by song length (longer than 5 minutes):
```bash
cloneheroer ./songs --length ">5:00"
```

Sort by year:
```bash
cloneheroer ./songs --sort year
```

Get count only:
```bash
cloneheroer ./songs --count
```

Fuzzy search song name:
```bash
cloneheroer ./songs --name "crow"
```

Write to file:
```bash
cloneheroer ./songs --output results.txt
```

## Flags

- `-o, --output string`: Write results to file instead of stdout
- `-c, --count`: Only return count of matching songs
- `-n, --name string`: Filter by song name (fuzzy matching)
- `-a, --artist string`: Filter by artist
- `-g, --genre string`: Filter by genre
- `--charter string`: Filter by charter
- `-y, --year int`: Filter by year
- `-l, --length string`: Filter by song length (e.g., '>5:00' or '<3:30')
- `-i, --instrument string`: Filter by instrument (guitar, drums, bass, etc.)
- `-s, --sort string`: Sort by field (name, artist, year, length, genre, charter)

## Cache

The tool caches song metadata in `$TMPDIR/cloneheroer/`. The cache is automatically invalidated when directory contents change based on file modification times.

## Song Format

The tool expects Clone Hero song directories with a `song.ini` file containing metadata in INI format:

```ini
[song]
name = Song Name
artist = Artist Name
album = Album Name
genre = Genre
year = 2023
charter = Charter Name
song_length = 180000
diff_guitar = 5
diff_drums = 4
```

## Additional Considerations

Some features that could be added:

1. **Multiple charter support**: Some charts have multiple contributors listed - currently only the first charter is parsed
2. **Playlist support**: Filter/sort by playlist information
3. **Difficulty filtering**: Filter by specific difficulty levels for instruments
4. **JSON/CSV output formats**: Alternative output formats for programmatic use
5. **Interactive mode**: TUI for browsing and filtering songs
6. **Export functionality**: Export filtered lists to playlists or other formats

