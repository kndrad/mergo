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
	"os"
	"path/filepath"

	"github.com/kndrad/mergo/internal/merge"
	"github.com/spf13/cobra"
)

var (
	InPath  string
	OutPath string
)

var mergoCmd = &cobra.Command{
	Use:   "mergo",
	Short: "Merge each Go package files into one",
	Long: `Mergo is a command-line tool that merges multiple Go files within a package into a single file.

It processes all non-test Go files in the specified input directory, combining them into a single file per package.
The tool preserves package structure, merges import statements, and maintains all declarations and functions.

Usage:
  mergo -p /path/to/input/directory -o /path/to/output/directory

This will process all Go packages in the input directory and create merged files in the output directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		in, err := filepath.Abs(filepath.Clean(InPath))
		if err != nil {
			return err
		}
		fmt.Println("Merging packages found at:", in)

		outDir := filepath.Dir(OutPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := merge.ManyPackages(in, OutPath); err != nil {
			return err
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
	mergoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	mergoCmd.Flags().StringVarP(&InPath, "path", "p", "", "Input path of the direcotyr")
	mergoCmd.Flags().StringVarP(&OutPath, "out", "o", ".", "Output path for merged files")
	mergoCmd.MarkFlagRequired("path")
	mergoCmd.MarkFlagRequired("out")
}
