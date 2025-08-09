# cgmerge

A command-line tool that merges source code files from a project into a single output, designed for use with Large Language Models (LLMs) and code analysis tools.

## Features

- **Multi-language support**: Currently supports Go and C# projects
- **Smart file filtering**: Automatically excludes test files, generated files, and build artifacts
- **Import/using consolidation**: Deduplicates and organizes import statements
- **Namespace management**: Consolidates C# files under a single namespace
- **Clipboard integration**: Optionally copy output directly to clipboard
- **File markers**: Adds source file markers for easy navigation in merged output

## Installation

```bash
go build -o cgmerge
```

## Usage

### Basic Usage

```bash
# Merge Go files in current directory
./cgmerge

# Merge C# files in current directory  
./cgmerge -lang csharp

# Merge files from specific directory
./cgmerge -dir /path/to/project

# Copy output to clipboard
./cgmerge -clipboard
```

### Command-line Options

- `-dir <path>`: Directory to scan for source files (default: current directory)
- `-lang <language>`: Language of files to merge - `go` or `csharp` (default: `go`)
- `-clipboard`: Copy merged output to clipboard

### Examples

```bash
# Merge Go project and copy to clipboard
./cgmerge -dir ./my-go-project -clipboard

# Merge C# console application
./cgmerge -lang csharp -dir ./my-csharp-app

# Merge current directory Go files
./cgmerge
```

## Language Support

### Go Projects
- Merges all `.go` files (excludes `*_test.go` and `*merged*.go`)
- Consolidates import statements
- Preserves package declarations
- Maintains proper Go formatting

### C# Projects  
- Merges all `.cs` files (excludes test files, `*.Designer.cs`, `*.g.cs`, `*merged.cs`)
- Processes `GlobalUsings.cs` if present
- Consolidates using statements
- Unifies all code under the namespace containing the `Main` method
- Handles both traditional and file-scoped namespaces

## Output Format

The merged output includes:
1. Consolidated import/using statements (sorted)
2. Package/namespace declaration
3. File content markers: `// --- /path/to/file.ext ---`
4. Source code with proper indentation

## Use Cases

- **LLM Code Analysis**: Provide entire codebase context to AI models
- **Code Reviews**: Generate single-file view of project for review
- **Documentation**: Create comprehensive code snapshots
- **Debugging**: Analyze complete program flow in one file

## Requirements

- Go 1.24 or later
- For clipboard functionality: Platform-specific clipboard support

## Dependencies

- `github.com/atotto/clipboard` - Cross-platform clipboard access
