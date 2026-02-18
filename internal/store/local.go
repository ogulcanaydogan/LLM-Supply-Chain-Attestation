package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func SaveLocal(srcPath, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := filepath.Base(srcPath)
	dst := filepath.Join(dir, name)
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return dst, nil
}

func EnsureDefaultAttestationDir() (string, error) {
	d := ".llmsa/attestations"
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", fmt.Errorf("create local store: %w", err)
	}
	return d, nil
}
