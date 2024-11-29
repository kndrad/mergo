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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/kndrad/mergo/internal/mergef"
	"github.com/spf13/cobra"
)

func Logger() *slog.Logger {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return l
}

var rootCmd = &cobra.Command{
	Use:   "mergo",
	Short: "Write Go module each Go package file to txt file",
	Long: `Mergo is a command-line tool that merges multiple Go files within a package into a single file.

It processes all non-test Go files in the specified input directory, combining them into a single file per package.
Command preserves package structure, merges import statements, and maintains all declarations and functions.

Usage:
  mergo -p /path/to/gomodule/directory -o /path/to/output/directory

Or you can also use it without any flags and it will process current module and output in this module with:
  mergo


This will process all Go packages and it's files in a directory and write to an output.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		modulePath := filepath.Clean(args[0])
		logger := Logger()

		if ok, err := mergef.IsGoMod(modulePath); !ok || err != nil {
			logger.Error("Failed to check if path is a Go module", "path", modulePath, "err", err)

			return fmt.Errorf("is module: %w", err)
		}

		files, err := mergef.WalkGoModule(modulePath)
		if err != nil {
			logger.Error("Failed to walk module", "err", err)

			return fmt.Errorf("walk module: %w", err)
		}

		// Open txt file
		timestamp := time.Now().Format("200601021504")
		outPath := filepath.Join(args[1], fmt.Sprintf("llm%s.txt", timestamp))
		txtf, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			logger.Error("Failed to open txt file", "path", outPath, "err", err)

			return fmt.Errorf("open file: %w", err)
		}
		defer txtf.Close()

		// Clear and reset to beginning
		if err := txtf.Truncate(0); err != nil {
			logger.Error("Failed to truncrate", "err", err)

			return fmt.Errorf("truncate: %w", err)
		}
		if _, err := txtf.Seek(0, 0); err != nil {
			logger.Error("Failed to seek 0 0", "err", err)

			return fmt.Errorf("seek: %w", err)
		}

		// Write each merged Go file content
		if err := mergef.WritePackages(files, txtf); err != nil {
			logger.Error("Failed to merge Go files:", "err", err)

			return fmt.Errorf("merge go files: %w", err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("module", "m", ".", "Go module path")
	rootCmd.Flags().StringP("out", "o", ".", "Output directory")
	rootCmd.MarkFlagsRequiredTogether("module", "out")
}
