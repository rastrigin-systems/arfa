package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestService_SelectWorkspace tests the workspace selection logic
func TestService_SelectWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		input         string // Simulated user input
		defaultPath   string
		expectedPath  string
		shouldError   bool
		errorContains string
	}{
		{
			name:         "Empty input uses default (current directory)",
			input:        "\n",
			defaultPath:  ".",
			expectedPath: ".", // Will be resolved to abs path
			shouldError:  false,
		},
		{
			name:         "Relative path is converted to absolute",
			input:        "../\n",
			defaultPath:  ".",
			expectedPath: "..", // Will be resolved
			shouldError:  false,
		},
		{
			name:         "Absolute path is used as-is",
			input:        "/tmp\n",
			defaultPath:  ".",
			expectedPath: "/tmp",
			shouldError:  false,
		},
		{
			name:          "Non-existent path returns error",
			input:         "/this/path/definitely/does/not/exist/12345\n",
			defaultPath:   ".",
			expectedPath:  "",
			shouldError:   true,
			errorContains: "does not exist",
		},
		{
			name:         "Empty string after trimming uses default",
			input:        "  \n",
			defaultPath:  ".",
			expectedPath: ".",
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService()

			// Simulate user input
			reader := strings.NewReader(tt.input)
			path, err := svc.SelectWorkspaceWithReader(tt.defaultPath, reader)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)

				// Convert expected to absolute for comparison
				expectedAbs, _ := filepath.Abs(tt.expectedPath)
				assert.Equal(t, expectedAbs, path)

				// Verify path is absolute
				assert.True(t, filepath.IsAbs(path), "Path should be absolute")
			}
		})
	}
}

// TestService_GetWorkspaceInfo tests workspace information gathering
func TestService_GetWorkspaceInfo(t *testing.T) {
	// Create a temporary directory with known files
	tempDir := t.TempDir()

	// Create some test files
	testFiles := []string{"file1.txt", "file2.go", "file3.md"}
	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		require.NoError(t, err)
		f.WriteString("test content")
		f.Close()
	}

	// Create a subdirectory with a file
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)
	f, err := os.Create(filepath.Join(subDir, "file4.txt"))
	require.NoError(t, err)
	f.WriteString("test content in subdir")
	f.Close()

	svc := NewService()
	info, err := svc.GetWorkspaceInfo(tempDir)

	require.NoError(t, err)
	assert.Equal(t, tempDir, info.Path)
	assert.Equal(t, 4, info.FileCount, "Should count all files including subdirectory")
	assert.Greater(t, info.TotalSize, int64(0), "Should have non-zero size")
	assert.True(t, info.Exists, "Directory should exist")
}

// TestService_GetWorkspaceInfo_NonExistent tests info for non-existent path
func TestService_GetWorkspaceInfo_NonExistent(t *testing.T) {
	svc := NewService()
	info, err := svc.GetWorkspaceInfo("/this/path/does/not/exist/12345")

	require.NoError(t, err) // Should not error, but info.Exists should be false
	assert.False(t, info.Exists)
	assert.Equal(t, 0, info.FileCount)
	assert.Equal(t, int64(0), info.TotalSize)
}

// TestService_ValidatePath tests path validation
func TestService_ValidatePath(t *testing.T) {
	svc := NewService()

	tests := []struct {
		name        string
		path        string
		expectValid bool
	}{
		{
			name:        "Current directory is valid",
			path:        ".",
			expectValid: true,
		},
		{
			name:        "Parent directory is valid",
			path:        "..",
			expectValid: true,
		},
		{
			name:        "Temp directory is valid",
			path:        "/tmp",
			expectValid: true,
		},
		{
			name:        "Non-existent path is invalid",
			path:        "/this/path/does/not/exist/12345",
			expectValid: false,
		},
		{
			name:        "Empty path is invalid",
			path:        "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidatePath(tt.path)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestService_FormatSize tests size formatting
func TestService_FormatSize(t *testing.T) {
	svc := NewService()

	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := svc.FormatSize(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}
