# xyzduck Copilot Instructions

## Project Overview

**xyzduck** is a cross-platform CLI tool for managing geospatial data in DuckDB databases. Built with Go 1.25.5, it provides database initialization with automatic spatial extension setup and self-updating capabilities.

**Key Dependencies:**
- `github.com/spf13/cobra` - CLI command framework
- `github.com/charmbracelet/bubbletea` & `bubbles` - Terminal UI (TUI)
- `github.com/duckdb/duckdb-go/v2` - DuckDB driver with spatial extension support
- `github.com/rhysd/go-github-selfupdate` - Self-updating mechanism

## Architecture & Code Organization

```
cmd/          - CLI command implementations (root.go, init.go, update.go)
src/
  database/   - DuckDB operations (connection, spatial extension setup)
  version/    - Version metadata injected at build time
main.go       - Entry point: delegates to cmd.Execute()
```

**Data Flow:**
1. User runs `xyzduck init [filename]` → `cmd/init.go` processes args/TUI input
2. `database.CreateOrOpenDatabase()` opens/creates .duckdb file via sql.Open("duckdb", path)
3. `database.InitSpatialExtension()` executes DuckDB SQL: `INSTALL spatial; LOAD spatial;`
4. User gets ready-to-use database with geospatial capabilities

**Critical Design Pattern:** All database operations in `src/database/duckdb.go` use `defer db.Close()` to ensure connections are released. Always follow this pattern when adding new database operations.

## CLI Command Structure

Commands are defined as `cobra.Command` in `cmd/` with the following pattern:

```go
var cmdCmd = &cobra.Command{
    Use: "command [args]",
    Short: "One-line description",
    Long: `Detailed multi-line description`,
    Args: cobra.MaximumNArgs(1),  // Validate arg count
    RunE: runCommand,             // Error-returning function
}

func init() {
    rootCmd.AddCommand(cmdCmd)   // Register with root
}
```

**When adding new commands:**
- Place command definition and `init()` in separate file under `cmd/`
- Use `RunE` for functions that can fail (returns error)
- Register with `rootCmd.AddCommand()` in `init()`
- Call `cmd.Execute()` in `main.go` automatically delegates

## TUI Pattern (cmd/init.go)

Interactive prompts use **Bubble Tea** with the Model-View-Update pattern:

```go
type filenameModel struct {
    textInput textinput.Model  // UI component
    submitted bool             // Track state
    err       error            // Display errors inline
}

func (m filenameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle input */ }
func (m filenameModel) View() string { /* render UI */ }
```

**Key behaviors to preserve:**
- `tea.KeyEnter` → submit with validation
- `tea.KeyCtrlC`/`tea.KeyEsc` → cancel
- Show errors inline in `View()` without exiting
- Use `tea.NewProgram(model).Run()` to launch

## Build & Release Workflow

**Local Development:**
```bash
go build -o xyzduck main.go      # Build binary
go run main.go                    # Run directly
go fmt ./...                      # Format code
go test ./...                     # Run tests
go mod tidy                       # Clean dependencies
```

**Release (via GoReleaser):**
```bash
goreleaser build --snapshot --clean        # Multi-platform binaries (Linux/Windows/macOS)
goreleaser release --snapshot --clean      # Full release (requires git tags)
goreleaser check                           # Validate config
```

GoReleaser configuration enables **CGO for DuckDB bindings** across platforms. Binaries include platform-specific DuckDB drivers (darwin-amd64, linux-x86_64, windows-amd64, etc.).

## Version Management

Version info is **injected at build time** via ldflags in GoReleaser. Package `src/version` stores:
- `Version` - Release tag (e.g., "v1.0.0")
- `Commit` - Git commit hash
- `Date` - Build timestamp

Access via:
```go
import "org.xyzmaps.xyzduck/src/version"
version.GetVersion()      // Returns "v1.0.0"
version.GetFullVersion()  // Returns "v1.0.0 (commit: abc123, built: 2024-12-28)"
```

**Never hardcode version strings** — always use the `version` package.

## Self-Update Mechanism (cmd/update.go)

The `update` command checks GitHub releases and applies binary updates:

```bash
xyzduck update --dry-run  # Check without updating
xyzduck update --yes      # Update without confirmation
```

Uses `github.com/rhysd/go-github-selfupdate` to detect latest release from `xyzmaps/xyzduck` repo and verify checksums. Follows pattern: detect → compare versions → prompt (unless `--yes`) → download → verify → replace binary.

## Extension Points

**For new geospatial features:**
- Add SQL execution functions to `src/database/duckdb.go` (follow defer pattern)
- DuckDB spatial extension is pre-loaded in initialized databases — use `EXECUTE` statements for queries

**For new CLI commands:**
- Create `cmd/newcommand.go` with `cobra.Command` and `init()` registration
- If command needs user input, implement Bubble Tea model in the same file
- Import from `src/database` and `src/version` packages as needed

**Module Path:** Always use `org.xyzmaps.xyzduck` for imports (defined in go.mod).

## Testing & Debugging

Currently minimal test coverage. When adding tests:
- Use `go test ./...` to run all packages
- Follow Go testing conventions: `*_test.go` files
- Test database operations against temporary .duckdb files (don't hardcode paths)

**Debugging database issues:**
- DuckDB driver errors come from `sql.Open()` and `db.Exec()` — wrap these in `fmt.Errorf()` for context
- Spatial extension failures indicate DuckDB version incompatibility — check bindings versions in go.mod
