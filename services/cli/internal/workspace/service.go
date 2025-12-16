package workspace

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Service handles workspace selection and validation
type Service struct{}

// NewService creates a new workspace service
func NewService() *Service {
	return &Service{}
}

// SelectWorkspace prompts the user to select a workspace directory
// If user presses Enter without input, uses defaultPath
func (s *Service) SelectWorkspace(defaultPath string) (string, error) {
	return s.SelectWorkspaceWithReader(defaultPath, os.Stdin)
}

// SelectWorkspaceWithReader is a testable version that accepts a custom reader
func (s *Service) SelectWorkspaceWithReader(defaultPath string, reader io.Reader) (string, error) {
	// Resolve default path to absolute
	defaultAbs, err := filepath.Abs(defaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve default path: %w", err)
	}

	// Show prompt with default
	fmt.Printf("Workspace [%s]: ", defaultAbs)

	// Read user input
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	// If empty, use default
	if input == "" {
		input = defaultPath
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(input)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate path exists
	if err := s.ValidatePath(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

// ValidatePath checks if a path exists and is accessible
func (s *Service) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Convert to absolute if relative
	absPath := path
	if !filepath.IsAbs(path) {
		var err error
		absPath, err = filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	return nil
}

// GetWorkspaceInfo gathers information about a workspace directory
func (s *Service) GetWorkspaceInfo(path string) (*Info, error) {
	info := &Info{
		Path:   path,
		Exists: false,
	}

	// Check if path exists
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return info, nil // Return info with Exists=false
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	info.Exists = true

	// If not a directory, just return size
	if !stat.IsDir() {
		info.FileCount = 1
		info.TotalSize = stat.Size()
		return info, nil
	}

	// Walk directory to count files and calculate size
	err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		if !fi.IsDir() {
			info.FileCount++
			info.TotalSize += fi.Size()
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return info, nil
}

// FormatSize formats a byte size into a human-readable string
func (s *Service) FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// DisplayWorkspaceInfo prints workspace information to the console
func (s *Service) DisplayWorkspaceInfo(info *Info) {
	if !info.Exists {
		fmt.Printf("Warning: Workspace not found: %s\n", info.Path)
		return
	}

	sizeStr := s.FormatSize(info.TotalSize)
	fmt.Printf("Workspace: %s (%s, %d files)\n", info.Path, sizeStr, info.FileCount)
}
