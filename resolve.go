package asset

import (
	"io"
	"os"
	"path/filepath"
)

type fileAsset string

func (a fileAsset) GetName() string {
	return filepath.Base(string(a))
}
func (a fileAsset) GetContent() (io.ReadCloser, error) {
	return os.Open(string(a))
}

func resolveFiles(patterns []string) <-chan Asset {
	result := make(chan Asset)
	go func() {
		for _, p := range patterns {
			if files, err := filepath.Glob(p); err == nil {
				for _, file := range files {
					ass := fileAsset(file)
					result <- ass
				}
			}
		}
		close(result)
	}()
	return result
}
