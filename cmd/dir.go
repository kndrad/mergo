/*
Copyright © 2024 Konrad Nowara

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Entries struct {
	dir   string
	paths []string
	excl  exclusions
}

func Walk(root string, excl exclusions) (*Entries, error) {
	entries := &Entries{
		dir:   root,
		paths: make([]string, 0),
		excl:  excl,
	}

	if err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		// Do not append excluded dirs and file by extensions (and dirs overall)
		if !isExcluded(path, excl) && !info.IsDir() {
			entries.paths = append(entries.paths, path)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}

	return entries, nil
}

// Returns most common extension which occurs in entries paths.
func (e *Entries) PopularExt() string {
	if e == nil {
		panic("entries cannot be nil")
	}
	if e.paths == nil {
		panic("entires paths cannot be nil")
	}
	ranking := make(map[string]int)
	for _, path := range e.paths {
		ranking[filepath.Ext(path)]++
	}
	type pair struct {
		ext   string
		count int
	}
	// Sort
	pairs := make([]pair, 0)
	for e, c := range ranking {
		pairs = append(pairs, pair{ext: e, count: c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})
	// Pick most popular
	s, _ := strings.CutPrefix(pairs[0].ext, ".")

	return s
}

type Entry struct {
	path string
	data []byte
}

func (e *Entry) Data() []byte {
	return e.data
}

func (e *Entry) read() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	e.data = data

	return nil
}

func (e *Entries) ReadAll() iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		entry := new(Entry)

		for _, path := range e.paths {
			entry.path = path

			if err := entry.read(); err != nil {
				yield(nil)
			}
			yield(entry)
		}
	}
}

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Combines content of files but only located within a directory.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := DefaultLogger()

		root, err := cmd.Flags().GetString("path")
		checkErr(logger, "Failed to get path flag string value", err)
		checkPathArg(root, logger)

		// Exclusions
		dirs, err := cmd.Flags().GetStringArray("exclude")
		checkErr(logger, "Failed to get exclude flag", err)

		exts, err := cmd.Flags().GetStringArray("exclude-ext")
		checkErr(logger, "Failed to get exclude-ext flag", err)

		// Load entries
		entries, err := Walk(root, exclusions{
			dirs: slices.Concat(dirs, defaultExcludedFiles()),
			exts: slices.Concat(exts, defaultExcludedExtensions()),
		})
		checkErr(logger, "Failed to walk to get entries", err)
		fmt.Printf("ENTRIES: %#v", entries)

		// Read all and append each file content
		contents := make([]string, 0)
		for entry := range entries.ReadAll() {
			if entry == nil {
				return errors.New("reading in entries failed")
			}
			builder := new(strings.Builder)
			if _, err := builder.WriteString(fmt.Sprintf("##### %s ######\n\n", entry.path)); err != nil {
				return fmt.Errorf("buffer write: %w", err)
			}
			if _, err := builder.Write(entry.Data()); err != nil {
				return fmt.Errorf("buffer write: %w", err)
			}
			if _, err := builder.WriteString("\n"); err != nil {
				return fmt.Errorf("buffer write: %w", err)
			}
			contents = append(contents, builder.String())
		}

		// Write each buffer to a file
		outpath, err := cmd.Flags().GetString("out")
		fmt.Println(outpath)
		checkErr(logger, "failed to get out string flag", err)
		checkPathArg(outpath, logger)

		timestamp := time.Now().Format("200601021504")
		txtf, err := os.OpenFile(
			filepath.Join(
				outpath,
				string(filepath.Separator),
				fmt.Sprintf("llm_%s_%s.txt", entries.PopularExt(), timestamp),
			),
			os.O_APPEND|os.O_CREATE|os.O_RDWR,
			0o600,
		)
		checkErr(logger, "Failed to open txt file", err)
		defer txtf.Close()

		if err := setupf(txtf, logger); err != nil {
			logger.Error("Failed to setup txt file", "err", err)

			return fmt.Errorf("setupf: %w", err)
		}

		// Write each appended file string content
		for _, content := range contents {
			if _, err := txtf.WriteString(content); err != nil {
				logger.Error("Failed to write string to txt file", "err", err)

				return fmt.Errorf("txtf write string: %w", err)
			}
		}

		logger.Info("Program completed successfully",
			slog.Int("code", 0),
		)

		return nil
	},
}

func setupf(f *os.File, logger *slog.Logger) error {
	if err := f.Truncate(0); err != nil {
		logger.Error("Failed to truncrate", "err", err)

		return fmt.Errorf("truncate: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		logger.Error("Failed to seek 0 0", "err", err)

		return fmt.Errorf("seek: %w", err)
	}

	return nil
}

func checkPathArg(path string, logger *slog.Logger) {
	path = filepath.Clean(path)

	stat, err := os.Stat(path)
	checkErr(logger, "Failed to get file info", err)

	if !stat.IsDir() {
		panic("path arg must be dir")
	}
}

func cutHome(path string) (string, error) {
	path = filepath.Clean(path)

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	s, found := strings.CutPrefix(path, dir)
	if !found {
		return path, nil
	}

	return s, nil
}

type exclusions struct {
	dirs []string
	exts []string
}

func isExcluded(path string, e exclusions) bool {
	path = filepath.Clean(path)
	dir := strings.Split(path, string(filepath.Separator))[0]
	ext := filepath.Ext(path)

	return slices.Contains(e.dirs, dir) || slices.Contains(e.exts, ext)
}

// Checks if err is not nil.
// When err is not nil, logs it with a message.
func checkErr(logger *slog.Logger, msg string, err error) {
	if err != nil {
		logger.Error(msg, "err", err)
		logger.Info("Program exit.")
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dirCmd)

	// Exclusions
	dirCmd.Flags().String("path", ".", "Directory path")
	dirCmd.Flags().String("out", ".", "Output directory")
	dirCmd.MarkFlagsRequiredTogether("path", "out")

	dirCmd.Flags().StringArray("exclude", defaultExcludedFiles(), "Exclude files")
	dirCmd.Flags().StringArray("exclude-ext", defaultExcludedExtensions(), "Exclude files by extensions")
}

func defaultExcludedFiles() []string {
	return []string{
		".git",
		".gitignore",
	}
}

func defaultExcludedExtensions() []string {
	return []string{
		".sum", // go.sum file
		".md",  // README files
		".mod",
	}
}
