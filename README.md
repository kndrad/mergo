# Mergo

Mergo is a command-line tool that merges multiple Go files within a package into a single file.
You can then copy paste the output into your fav LLM without unnecesary ctrl+c ctrl+v each file.

Created for learning purposes and to explore Go 1.23's new iterator features.

## Details

- Combines all non-test Go files in a package
- Preserves package structure and imports
- Handles multiple packages in a directory
- Removes duplicate imports

## Installation

```bash
go install github.com/kndrad/mergo@latest
```

## Usage

```bash
mergo -p /path/to/input/directory -o /path/to/output/directory
```

- `-p, --path`: Input directory containing Go packages (required)
- `-o, --out`: Output directory for merged files (required)

## Example

```bash
mergo -p ./myproject -o ./merged
```

This command will process all Go packages in `./myproject` and create merged `.go` files in the `./merged` directory.

## License

MIT License - see the [LICENSE](LICENSE) file for details.