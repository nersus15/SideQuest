// Package store persists small JSON documents atomically.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// File is a JSON document on disk.
type File struct {
	Path string
}

// Load decodes the file into v. A missing file returns (false, nil).
// A corrupt file is moved aside so the app can start fresh instead of
// crashing, and (false, err) is returned for logging.
func (f File) Load(v any) (bool, error) {
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, v); err != nil {
		backup := fmt.Sprintf("%s.corrupt-%d", f.Path, time.Now().Unix())
		_ = os.Rename(f.Path, backup)
		return false, fmt.Errorf("state file was unreadable, moved to %s: %w", backup, err)
	}
	return true, nil
}

// Save writes v via a temp file and rename, so a crash mid-write never
// leaves a half-written state file behind.
func (f File) Save(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(f.Path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, f.Path)
}
