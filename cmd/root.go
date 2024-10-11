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

	"github.com/kndrad/mergo/internal/mergo"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var (
	modPath string
	outPath string
)

var mergoCmd = &cobra.Command{
	Use:   "mergo",
	Short: "Merge each Go package files into one",
	Long: `Mergo is a command-line tool that merges multiple Go files within a package into a single file.

It processes all non-test Go files in the specified input directory, combining them into a single file per package.
The tool preserves package structure, merges import statements, and maintains all declarations and functions.

Usage:
  mergo -p /path/to/gomodule/directory -o /path/to/output/directory

This will process all Go packages and it's files in a directory and write to an output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		modPath := filepath.Clean(modPath)
		logger.Info("mergo:", "modPath", modPath)

		if ok, err := mergo.IsModule(modPath); !ok || err != nil {
			logger.Error("mergoCmd: not a Go module", "modPath", modPath)

			return fmt.Errorf("mergoCmd: %w", err)
		}

		files, err := mergo.ModulePackageFiles(modPath)
		if err != nil {
			logger.Error("mergoCmd:", "err", err)

			return fmt.Errorf("mergoCmd: %w", err)
		}

		outPath = filepath.Clean(outPath) + string(filepath.Separator) + "llm_input.txt"
		fmt.Println(outPath)
		outFile, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		defer outFile.Close()

		// Clear outFile and reset to beginning
		if err := outFile.Truncate(0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("cmd: %w", err)
		}
		if _, err := outFile.Seek(0, 0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("cmd: %w", err)
		}
		if err := mergo.ProcessFiles(files, outFile); err != nil {
			logger.Error("mergoCmd:", "err", err)

			return fmt.Errorf("mergoCmd: %w", err)
		}
		fmt.Println("Merging done")

		return nil
	},
}

func Execute() {
	err := mergoCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	mergoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	mergoCmd.Flags().StringVarP(&modPath, "path", "p", "", "Path of Go module")
	mergoCmd.Flags().StringVarP(&outPath, "out", "o", ".", "Output path")

	mergoCmd.MarkFlagsRequiredTogether("path", "out")
}
