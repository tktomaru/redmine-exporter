# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This repository contains two implementations of a Redmine ticket exporter:

1. **VBA Implementation** (`vb/`) - Excel macro for exporting tickets to Excel tables
2. **Go Implementation** (`go/`) - Cross-platform CLI tool supporting multiple output formats

Both implementations share the same INI configuration file format (`redmine.config`) for compatibility.

## Build and Test Commands

### Go Implementation

All commands should be run from the `go/` directory:

```bash
# Build single binary for current platform
make build

# Cross-compile for Linux, macOS, Windows
make build-all

# Run all tests with verbose output
make test

# Run tests with coverage report (opens HTML report)
make test-coverage

# Run a single test
go test -v ./internal/config -run TestLoadConfig

# Run tests for a specific package
go test -v ./internal/formatter

# Clean build artifacts
make clean

# Build and run (requires ../redmine.config)
make run
```

### Running the CLI

```bash
# From go/ directory
./bin/redmine-exporter -o output.xlsx
./bin/redmine-exporter -c custom.config -o output.md
./bin/redmine-exporter -h  # Show help
./bin/redmine-exporter -v  # Show version
```

### Docker Environment

```bash
# Start local Redmine (requires .env file)
docker-compose -f docker-compose-redmine.yml up -d

# Stop Redmine
docker-compose -f docker-compose-redmine.yml down

# Access: http://localhost:3000 (admin/admin)
```

Before running Docker, copy `.env.sample` to `.env` and configure:
```bash
cp .env.sample .env
# Edit .env with your values, especially REDMINE_SECRET_KEY_BASE
```

## Architecture

### High-Level Data Flow

```
Config File (INI) → Redmine API (paginated) → Processor → Formatter → Output File
```

1. **Config Loading** (`internal/config`): Reads INI file with VBA-compatible Pattern1, Pattern2... format
2. **API Client** (`internal/redmine`): Fetches all issues with automatic pagination (100 items/page)
3. **Processor** (`internal/processor`): Builds parent-child relationships, cleans titles, extracts summaries
4. **Formatter** (`internal/formatter`): Extension-based format detection, outputs to Text/Markdown/Excel

### Key Architectural Decisions

**Parent-Child Relationship Building**
- Only issues with a parent are output (standalone issues are ignored)
- Parent issues without children are also ignored
- Structure: `Process()` builds a map by ID, then iterates to build parent.Children arrays

**Title Cleaning Pattern**
- Uses VBA-compatible regex patterns (Pattern1, Pattern2, ...)
- Invalid regex patterns are silently skipped (VBA compatibility)
- Patterns are applied sequentially to each title

**Summary Extraction**
- Priority 1: Text between `[要約]...[/要約]` tags
- Priority 2: First non-empty line of description
- Both leading/trailing whitespace are trimmed

**Date Handling**
- Custom `Date` type with embedded `time.Time`
- Implements custom `UnmarshalJSON` to handle null, empty strings, and "2006-01-02" format
- `Format()` returns "YYYY/MM/DD" or "----/--/--" for zero values

**Formatter Interface**
- Strategy pattern: Each formatter implements `Format(roots []*Issue, w io.Writer) error`
- Extension-based auto-detection in `DetectFormatter(filename string)`
- ExcelFormatter is special: requires filename in constructor for write-to-disk

### VBA Compatibility

The Go implementation maintains compatibility with VBA version:

1. **Config File**: Same INI structure, Pattern1/Pattern2 numbering
2. **Output Format**: Text and Excel formatters produce VBA-compatible output
3. **Behavior**: Invalid regex patterns are skipped (not errors)
4. **API Usage**: Same URL construction, same pagination limit (100)

### Important File Relationships

- `main.go` orchestrates: config → client → processor → formatter
- `models.go` defines all JSON structures matching Redmine API
- `processor.go` is stateful: stores compiled regex patterns
- All formatters are stateless except ExcelFormatter (needs filename)

## Configuration File Format

Required sections:

```ini
[Redmine]
BaseUrl=https://redmine.example.com
ApiKey=YOUR_API_KEY
FilterUrl=/issues.json?project_id=1&status_id=*

[TitleCleaning]
Pattern1=^\[.*?\]\s*     # Remove [WIP], [Done], etc.
Pattern2=\s*\(.*?\)$     # Remove (suffix text)
# Pattern3, Pattern4... continue numbering
```

**Critical**:
- FilterUrl must end with `.json`
- Pattern numbering must be sequential (Pattern1, Pattern2, ...)
- Missing patterns break the loop in config loader

## Testing Patterns

All tests use table-driven testing:

```go
tests := []struct {
    name    string
    input   X
    want    Y
    wantErr bool
}{
    {name: "case1", input: ..., want: ..., wantErr: false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

**Coverage Notes**:
- config: 100% (complete)
- processor: 97.7% (near complete)
- formatter: 49.2% (Excel writing not tested, text/markdown covered)
- redmine: 22.2% (HTTP client not tested, models fully tested)

## Output Format Examples

**Text** (VBA-compatible):
```
■親タスクA
・子タスクB 【進行中】 2026/01/02-2025/12/31 担当: 佐藤
⇒要約テキスト
```

**Markdown**:
```markdown
# 親タスクA

- **子タスクB** [進行中] 2026/01/02-2025/12/31 担当: 佐藤
  > 要約テキスト
```

**Excel**: Table with columns: 親タスク | タスク名 | ステータス | 開始日 | 終了日 | 担当者 | 要約

## Common Patterns

### Adding a New Output Format

1. Create `internal/formatter/newformat.go`
2. Implement `Format(roots []*Issue, w io.Writer) error`
3. Add case to `DetectFormatter()` in `formatter.go`
4. Add tests in `formatter_test.go`

### Adding Config Parameters

1. Add field to struct in `internal/config/config.go`
2. Load in `LoadConfig()` using `cfg.Section("...").Key("...").String()`
3. Add to `Validate()` if required
4. Update tests in `config_test.go`

### Modifying Issue Processing

All processing logic is in `processor.Process()`. To add new transformations:
1. Add fields to `Issue` struct in `redmine/models.go` with `json:"-"` tag
2. Populate fields in `Process()` loop
3. Use in formatters

## Dependencies

- `gopkg.in/ini.v1` - INI config parsing
- `github.com/xuri/excelize/v2` - Excel file generation
- Standard library: `net/http`, `encoding/json`, `regexp`, `flag`

## Security Notes

- `redmine.config` and `.env` are gitignored (contain API keys)
- Use `redmine.config.sample` and `.env.sample` as templates
- Docker volumes in `docker/` directory are also gitignored
