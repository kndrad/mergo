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
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Combines content of files but only located within a directory.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := Logger()
		path := filepath.Clean(args[0])

		// Check if path is a directory
		stat, err := os.Stat(path)
		checkErr(logger, "Failed to get file info", err)

		if !stat.IsDir() {
			logger.Error("Path must be a directory", slog.String("path", path))

			return fmt.Errorf("is dir: %w", err)
		}

		// Exclude certain dirs
		exclusions, err := cmd.Flags().GetStringArray("exclude")
		checkErr(logger, "Failed to get exlude flag", err)
		fmt.Println(exclusions)

		// Walk dir
		if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			return nil
		}); err != nil {
			logger.Error("Failed to walk dir", slog.String("path", path))

			return fmt.Errorf("walk dir: %w", err)
		}

		logger.Info("Program completed successfully",
			slog.Int("code", 0),
		)

		return nil
	},
}

// Checks if err is not nil.
// When err is not nil, logs it with a message.
func checkErr(l *slog.Logger, msg string, err error) {
	if err != nil {
		l.Error(msg, "err", err)
		l.Info("Program exit.")
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dirCmd)

	// Exclusions
	dirCmd.Flags().StringP("path", "p", ".", "Directory path")
	dirCmd.Flags().StringP("out", "o", ".", "Output directory")
	dirCmd.MarkFlagsRequiredTogether("path", "out")

	dirCmd.Flags().StringArray("exclude", []string{}, "Exclude file extensions from merging")
}
