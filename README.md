# xyzduck

[![Release](https://img.shields.io/github/v/release/xyzmaps/xyzduck)](https://github.com/xyzmaps/xyzduck/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/xyzmaps/xyzduck/release.yml)](https://github.com/xyzmaps/xyzduck/actions)
[![Go Version](https://img.shields.io/badge/go-1.25.5-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/xyzmaps/xyzduck)](https://goreportcard.com/report/github.com/xyzmaps/xyzduck)

A CLI tool for working with geospatial mapping data in DuckDB databases.

## Features

- **Database Initialization**: Create or open DuckDB databases with spatial extension support
- **Interactive TUI**: User-friendly prompts powered by Bubble Tea
- **Self-Updating**: Built-in command to update to the latest release
- **Geospatial Ready**: Automatic installation of DuckDB's spatial extension for GIS operations

## Installation

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/xyzmaps/xyzduck/releases).

**Linux:**
```bash
# Download and extract
wget https://github.com/xyzmaps/xyzduck/releases/latest/download/xyzduck_Linux_x86_64.tar.gz
tar -xzf xyzduck_Linux_x86_64.tar.gz

# Move to PATH
sudo mv xyzduck /usr/local/bin/
```

**macOS:**
```bash
# Download and extract
wget https://github.com/xyzmaps/xyzduck/releases/latest/download/xyzduck_Darwin_x86_64.tar.gz
tar -xzf xyzduck_Darwin_x86_64.tar.gz

# Move to PATH
sudo mv xyzduck /usr/local/bin/
```

### Build from Source

Requirements:
- Go 1.25.5 or later
- C compiler (gcc or clang) for CGO support

```bash
git clone https://github.com/xyzmaps/xyzduck.git
cd xyzduck
go build -o xyzduck main.go
```

## macOS: Unblocking the unsigned `xyzduck` CLI executable

If you downloaded a prebuilt `xyzduck` binary and macOS prevents it from running because it is unsigned, you can use one of the methods below to allow it to run. Always verify the binary's integrity and that you trust the source before removing security controls.

Recommended steps (safe and minimal)
1. Make the binary executable (if it isn't already):
   ```bash
   chmod +x /path/to/xyzduck
   ```
2. Remove the quarantine attribute that Gatekeeper sets when a file is downloaded:
   ```bash
   xattr -d com.apple.quarantine /path/to/xyzduck
   ```
   If you need to remove the quarantine attribute recursively on a directory:
   ```bash
   xattr -r -d com.apple.quarantine /path/to/directory
   ```
3. Move the binary into a standard location (optional):
   ```bash
   sudo mv /path/to/xyzduck /usr/local/bin/xyzduck
   ```

Alternative (GUI) — allow once via Finder
- In Finder, right-click (or Control-click) the `xyzduck` file and choose "Open". macOS will show a warning; click "Open" to allow it to run this one time. This creates a one-time exception in Gatekeeper.

## Usage

### Initialize a Database

Create a new DuckDB database with spatial extension:

```bash
# With filename argument
xyzduck init mydata
# Creates: mydata.duckdb

# Interactive mode (prompts for filename)
xyzduck init
```

The `init` command:
- Creates a new `.duckdb` file or opens an existing one
- Automatically installs the DuckDB spatial extension
- Loads the spatial extension for immediate use
- Is idempotent - safe to run multiple times on the same database

### Update xyzduck

Keep xyzduck up to date with the latest release:

```bash
# Check for updates (dry run)
xyzduck update --dry-run

# Update with confirmation prompt
xyzduck update

# Update without confirmation
xyzduck update --yes
```

### Version Information

```bash
xyzduck --version
```

### Help

```bash
# General help
xyzduck --help

# Command-specific help
xyzduck init --help
xyzduck update --help
```

## DuckDB Spatial Extension

The spatial extension provides geospatial functionality including:
- Geometry types (POINT, LINESTRING, POLYGON, etc.)
- Spatial operations (intersects, contains, distance, etc.)
- Coordinate reference system transformations
- GeoJSON and WKT support

Example usage (for now with DuckDB CLI, this will be available in the xyzduck CLI soon):
```sql
-- After running: xyzduck init geodata

-- Create a table with spatial data
CREATE TABLE locations (
    id INTEGER,
    name VARCHAR,
    geom GEOMETRY
);

-- Insert spatial data
INSERT INTO locations VALUES
    (1, 'Point A', ST_Point(0, 0)),
    (2, 'Point B', ST_Point(1, 1));

-- Query spatial data
SELECT name, ST_AsText(geom) FROM locations;
```

## Supported Platforms

- Linux (amd64, arm64)
- macOS / Darwin (amd64, arm64)
- Windows (amd64 only, unfortunately duckdb-go doesn't support arm64 yet)

## Development

### Project Structure

```
xyzduck/
├── cmd/              # Command implementations
│   ├── root.go      # Root command
│   ├── init.go      # Database initialization
│   └── update.go    # Self-update command
├── src/
│   ├── database/    # DuckDB operations
│   └── version/     # Version information
├── main.go          # Entry point
└── .goreleaser.yaml # Release configuration
```

### Building Releases

Releases are automatically built and published via GitHub Actions when a new tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

For local testing:
```bash
goreleaser build --snapshot --clean
```

### Dependencies

- [DuckDB Go Driver](https://github.com/duckdb/duckdb-go) - Official DuckDB driver with CGO
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- [DuckDB Documentation](https://duckdb.org/docs/)
- [DuckDB Spatial Extension](https://duckdb.org/docs/extensions/spatial.html)
- [GitHub Repository](https://github.com/xyzmaps/xyzduck)
