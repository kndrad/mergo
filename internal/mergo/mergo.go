package mergo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoPackagesFound is returned when no Go files are found in the specified directory.
var ErrNoPackagesFound = errors.New("no go package files found")

// ProcessFiles processes all Go packages found in the Go module directory and merges all files found in packages.
func ProcessFiles(files map[string]*ast.File, w io.Writer) error {
	fset := token.NewFileSet()

	var b strings.Builder

	for name, file := range files {
		b.WriteString("################" + "\n")
		b.WriteString("PACKAGE" + " " + name + "\n")
		b.WriteString("################" + "\n\n")
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
			b.WriteString("\n\n")
		}
		if _, err := w.Write([]byte(b.String())); err != nil {
			return fmt.Errorf("mergo: %w", err)
		}
		b.Reset()
	}

	return nil
}

func ModulePackageFiles(path string) (map[string]*ast.File, error) {
	// Walk the file tree.
	root := filepath.Clean(path)
	fmt.Println(root)

	fset := token.NewFileSet()
	files := make(map[string]*ast.File)

	var (
		// Does not include test files.
		filter = func(fi fs.FileInfo) bool {
			filename := fi.Name()
			isTestFile := !strings.HasSuffix(filename, "_test.go")
			isModFile := !strings.HasPrefix(filename, "go.mod")

			return isTestFile || isModFile
		}
		// Adds map of package name -> package AST to the objects slice
		walk = func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)

				return err
			}

			dir := filepath.Dir(path)

			pkgs, err := parser.ParseDir(fset, dir, filter, 0)
			if err != nil {
				return fmt.Errorf("mergo: %w", err)
			}
			for pkg := range maps.Values(pkgs) {
				files[pkg.Name] = ast.MergePackageFiles(pkg, ast.FilterImportDuplicates|ast.FilterUnassociatedComments)
			}

			return nil
		}
	)

	err := filepath.Walk(root, walk)
	if err != nil {
		return nil, fmt.Errorf("mergo: %w", err)
	}

	return files, nil
}

var ErrInvalidModule = errors.New("invalid Go Module")

func IsModule(path string) (bool, error) {
	if !checkDir(path) {
		return false, ErrInvalidModule
	}
	name := filepath.Join(filepath.Clean(path), "go.mod")
	f, err := os.OpenFile(name, os.O_RDONLY, 0o666)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("mergo: %w", err)
	}
	defer f.Close()

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
