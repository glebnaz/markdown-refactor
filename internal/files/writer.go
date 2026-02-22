package files

import (
	"fmt"
	"os"
)

// WriteFile writes content to the specified file path, replacing its contents.
// It preserves the original file's permissions.
func WriteFile(path, content string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), info.Mode()); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// ReadFile reads the entire contents of a file as a string.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(data), nil
}
