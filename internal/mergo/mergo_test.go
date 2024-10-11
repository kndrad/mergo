package mergo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kndrad/mergo/internal/mergo"
)

const (
	TestDataDir = "testdata"
	TestTmpFile = "tmpfile"
	TestTmpDir  = "tmp"
	GoVersion   = "1.23.2"
)

func Test_ModulePkgFiles(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Logf("Test_ModulePkgFiles: wd: %s", wd)
	tmpDirPath := filepath.Join(wd, TestDataDir)
	if !IsValidTestSubPath(t, tmpDirPath) {
		t.Error("not a valid test subpath", tmpDirPath)
	}
	t.Logf("Test_ModulePkgFiles: tempDirPath %s", tmpDirPath)

	tmpDir, err := os.MkdirTemp(tmpDirPath, TestTmpDir)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	t.Logf("Test_ModulePkgFiles: tempDir %#v", tmpDir)

	tmpModFile, err := os.CreateTemp(tmpDir, "go.mod")
	require.NoError(t, err)
	defer os.Remove(tmpModFile.Name())
	t.Logf("Test_ModulePkgFiles: created tmpModFile: %#v", tmpModFile.Name())

	require.NoError(t, err)
	topDecl := "module github.com/kndrad/tmpmod\n\n" + "go " + GoVersion + "\n\n"
	if _, err := tmpModFile.WriteString(topDecl); err != nil {
		t.Logf("Test_ModulePkgFiles: top go.mod declaration err: %v", err)
		t.FailNow()
	}
	t.Logf("Test_ModulePkgFiles: wrote go.mod declaration")

	filesTotal := 4
	tmpFilenames := make([]string, 0, filesTotal)
	for range filesTotal {
		tmpFile, err := os.CreateTemp(tmpDir, TestTmpFile+"*.go")
		require.NoError(t, err)
		tmpFilenames = append(tmpFilenames, tmpFile.Name())
		defer os.Remove(tmpFile.Name())
		t.Logf("Test_ModulePkgFiles: created tmpFile: %#v", tmpFile.Name())
	}

	tmpPkgFilesTotal := filesTotal - 1
	tmpPkgName := "mergetmp"
	for _, tmpFilename := range tmpFilenames[:tmpPkgFilesTotal] {
		tmpFile, err := os.OpenFile(tmpFilename, os.O_WRONLY, os.ModePerm)
		defer func() {
			if err := tmpFile.Close(); err != nil {
				t.Logf("Test_ModulePkgFiles err closing tmpFile: %v", err)
				t.FailNow()
			}
		}()
		t.Logf("Test_ModulePkgFiles opened file: %s", tmpFile.Name())
		require.NoError(t, err)

		topDecl := "package " + tmpPkgName
		if _, err := tmpFile.WriteString(topDecl); err != nil {
			t.Logf("Test_ModulePkgFiles err: %v", err)
			t.FailNow()
		}
		t.Logf("Test_ModulePkgFiles wrote top declaration: %s", topDecl)
	}

	tmpPkgDirPath := filepath.Join(tmpDir, tmpPkgName)
	if err := os.Mkdir(tmpPkgDirPath, 0o777); err != nil {
		t.FailNow()
		t.Logf("test_ModulePkgFiles err: %v", err)
	}
	t.Logf("Test_ModulePkgFiles created tmpPkg dir at: %v", tmpPkgDirPath)

	for i, tmpFilename := range tmpFilenames[:tmpPkgFilesTotal] {
		if file, err := os.OpenFile(tmpFilename, os.O_RDONLY, 0o666); err == nil {
			t.Logf("Test_ModulePkgFiles closing file: %s", tmpFilename)
			if err := file.Close(); err != nil {
				t.FailNow()
				t.Logf("Test_ModulePkgFiles err: %v", err)
			}
		}
		tmpPkgFilename := filepath.Join(tmpPkgDirPath, filepath.Base(tmpFilename))
		if err := os.Rename(tmpFilename, tmpPkgFilename); err != nil {
			t.Logf("Test_ModulePkgFiles err: %v", err)
			t.FailNow()
		}

		tmpFilenames[i] = tmpPkgFilename

		t.Logf("Test_ModulePkgFiles: moved file from %s to %s", tmpFilename, tmpPkgFilename)
	}

	tmpMainPkgName := "main"
	for _, tmpFilename := range tmpFilenames[tmpPkgFilesTotal:] {
		tmpFile, err := os.OpenFile(tmpFilename, os.O_WRONLY, os.ModeAppend)
		t.Logf("Test_ModulePkgFiles opened file: %s", tmpFile.Name())
		require.NoError(t, err)
		if _, err := tmpFile.WriteString("package " + tmpMainPkgName); err != nil {
			t.Logf("Test_ModulePkgFiles err: %v", err)
			t.FailNow()
		}
		t.Logf("Test_ModulePkgFiles wrote top declaration: %v", topDecl)
	}

	path := tmpDir
	files, err := mergo.ModulePkgFiles(path)

	fmt.Printf("files: %#v\n", files)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	// Total amount of merged files must equal 2
	assert.Len(t, files, 2)
}

func Test_IsModule(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("Test_IsModule: wd: %s", wd)

	tmpModDirPath := filepath.Join(wd, TestDataDir)
	t.Logf("Test_IsModule: tempDirPath %s", tmpModDirPath)

	// Create a temporary directiory for output files
	tmpModFile, err := os.Create(filepath.Join(tmpModDirPath, "go.mod"))
	require.NoError(t, err)
	t.Logf("Test_IsModule: created %#v", tmpModFile.Name())

	testcases := map[string]struct {
		path     string
		expected bool
		mustErr  bool
	}{
		"valid_module_path": {
			path:     filepath.Dir(tmpModFile.Name()),
			expected: true,
			mustErr:  false,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			isMod, err := mergo.IsModule(tc.path)
			if tc.mustErr {
				require.Error(t, err)
			}
			require.NoError(t, err)

			t.Logf("Test_IsModule: testing path: %s", tc.path)
			require.Equal(t, tc.expected, isMod)
		})
	}

	t.Cleanup(func() {
		os.Remove(tmpModFile.Name())
	})
}

func IsValidTestSubPath(t *testing.T, path string) bool {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	return IsSubPath(wd, path)
}

// IsSubPath checks if the filePath is a subpath of the base path.
func IsSubPath(basePath, filePath string) bool {
	rel, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
