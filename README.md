# mergo

mergo is a command-line tool that merges multiple Go files found in a Go module for easy output. You can then copy-paste the output into your favorite LLM without unnecessary ctrl+c ctrl+v for each Go file.

Created for learning purposes.

Go version: go1.23.2

## Details

- Combines all non-test Go files in a package
- Preserves package structure and imports
- Handles multiple packages in a directory
- Removes duplicate imports

## Installation

To install mergo, follow these steps:

1. Ensure you have Go 1.23.2 or later installed on your system.
2. Clone the repository:
   ```
   git clone https://github.com/kndrad/mergo.git
   ```
3. Navigate to the project directory:
   ```
   cd mergo
   ```
4. Build the project:
   ```
   go build
   ```
5. (Optional) Add the built binary to your PATH for easy access.

## Usage

```bash
mergo -p /path/to/input/directory -o /path/to/output/directory
```

- `-p, --path`: Input directory containing Go packages (required)
- `-o, --out`: Output directory for merged files (required)

## Example

```bash
mergo -p ./myproject -o ./out
```

This command will process all Go packages in `./myproject` and create a merged `llm_input.txt` file in the `./out` directory.

## Version
go version go1.23.2

## License

MIT License - see the [LICENSE](LICENSE) file for details.
