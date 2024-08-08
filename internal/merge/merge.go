package merge

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrNoFilesFound is returned when no Go files are found in the specified directory
	ErrNoFilesFound = errors.New("go files not found in specified dir path")
)

// ManyPackages processes all Go packages in the input directory and merges files within each package
func ManyPackages(in, out string) error {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, in, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return ErrNoFilesFound
	}

	files := make(map[string]*ast.File)

	for pkg := range maps.Values(packages) {
		files[pkg.Name] = ast.MergePackageFiles(pkg, ast.FilterImportDuplicates|ast.FilterUnassociatedComments)
	}

	for name, file := range files {
		var b strings.Builder

		b.WriteString(fmt.Sprintf("package %s\n\nimport (\n", name))
		for _, spec := range file.Imports {
			b.WriteString(fmt.Sprintf("\t"))
			if err := format.Node(&b, fset, spec); err != nil {
				return err
			}
			b.WriteString("\n")
		}
		b.WriteString(")\n\n")

		for _, d := range file.Decls {
			if ast.FilterDecl(d, func(s string) bool {
				return s != token.IMPORT.String()
			}) {
				b.WriteString("\n")
				if err := format.Node(&b, fset, d); err != nil {
					return err
				}
			}
			b.WriteString("\n")
		}

		if err := os.WriteFile(filepath.Join(filepath.Dir(out), name+".txt"), []byte(b.String()), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", out, err)
		}
	}

	return nil
}

// isTestFunc checks if a function name indicates a test function
func isTestFunc(name string) bool {
	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example")
}
