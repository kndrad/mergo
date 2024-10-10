package mergo

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

// ErrNoFilesFound is returned when no Go files are found in the specified directory.
var ErrNoFilesFound = errors.New("go files not found in specified dir path")

// Module processes all Go packages found in the Go module directory and merges all files found in packages.
func Module(path, out string) error {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, path, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return fmt.Errorf("mergo: %w", err)
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
			b.WriteString("\t")
			if err := format.Node(&b, fset, spec); err != nil {
				return fmt.Errorf("mergo: %w", err)
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
					return fmt.Errorf("mergo: %w", err)
				}
			}
			b.WriteString("\n")
		}

		if err := os.WriteFile(filepath.Join(filepath.Dir(out), name+".txt"), []byte(b.String()), 0o600); err != nil {
			return fmt.Errorf("mergo: failed to write file %s: %w", out, err)
		}
	}

	return nil
}

func ModulePkgFiles(path string) (map[string]*ast.File, error) {
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, path, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("mergo: %w", err)
	}
	if len(packages) == 0 {
		return nil, ErrNoFilesFound
	}

	files := make(map[string]*ast.File)

	for pkg := range maps.Values(packages) {
		files[pkg.Name] = ast.MergePackageFiles(pkg, ast.FilterImportDuplicates|ast.FilterUnassociatedComments)
	}

	return files, nil
}

var ErrInvalidModule = errors.New("invalid Go Module")

func IsModule(path string) (bool, error) {
	if !checkDir(path) {
		return false, ErrInvalidModule
	}
	name := filepath.Join(filepath.Clean(path), "go.mod")
	modFile, err := os.OpenFile(name, os.O_RDONLY, 0o666)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("isModule: %w", err)
	}
	defer modFile.Close()

	return true, nil
}

func checkDir(dir string) bool {
	info, err := os.Stat(filepath.Dir(dir))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	if !info.IsDir() {
		return false
	}

	return true
}
