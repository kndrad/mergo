# In progress
go version go1.23.2

# mergo

mergo is a command-line tool that merges multiple Go files found in go module for an easy output.
You can then copy paste the output into your fav LLM without unnecesary ctrl+c ctrl+v for each go file.

Created for learning purposes.

## Details

- Combines all non-test Go files in a package
- Preserves package structure and imports
- Handles multiple packages in a directory
- Removes duplicate imports

## Installation


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

This command will process all Go packages in `./myproject` and create merged `.go` files in the `./merged` directory.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
