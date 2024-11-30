# mergo

A CLI tool that combines Go files from a module or directory into a single file for easy sharing with LLMs.

## Features

- Merges Go files while preserving package structure
- Excludes test files and unwanted directories/extensions
- Supports multiple packages
- Handles imports deduplication
- Outputs in a format suitable for LLM prompts

## Installation

```bash
go install github.com/kndrad/mergo@latest
```

Or build from source:
```bash
git clone https://github.com/kndrad/mergo.git
cd mergo
go build
```

## Usage

Merge Go module files:
```bash
mergo /path/to/module /path/to/output
```

Merge directory files:
```bash
mergo dir --path=/path/to/dir --out=/path/to/output
```

### Options

- `--exclude`: Exclude directories (default: .git, .gitignore)
- `--exclude-ext`: Exclude file extensions (default: .sum, .md, .mod)

## Requirements

- Go 1.23.3 or later

## License

MIT License - see [LICENSE](LICENSE)
