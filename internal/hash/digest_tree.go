package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TreeEntry struct {
	Path     string
	Digest   string
	Size     int64
	Manifest string
}

func DigestTree(root string) (digest string, manifest string, entries []TreeEntry, err error) {
	entries = make([]TreeEntry, 0)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		norm := filepath.ToSlash(rel)
		fileDigest, size, err := DigestFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, TreeEntry{Path: norm, Digest: fileDigest, Size: size})
		return nil
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("walk tree %s: %w", root, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	var sb strings.Builder
	for i := range entries {
		line := fmt.Sprintf("%s\x00%s\x00%d\n", entries[i].Path, entries[i].Digest, entries[i].Size)
		entries[i].Manifest = line
		sb.WriteString(line)
	}

	manifest = sb.String()
	h := sha256.Sum256([]byte(manifest))
	return "sha256:" + hex.EncodeToString(h[:]), manifest, entries, nil
}

func DigestBytes(raw []byte) string {
	h := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(h[:])
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
