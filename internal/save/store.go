package save

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Store persists raw save bytes. Implementations: MemStore (tests),
// FileStore (desktop), LocalStore (browser localStorage, js/wasm only).
type Store interface {
	Load() ([]byte, error) // ErrNoSave if nothing saved
	Save(b []byte) error
	Clear() error
	// Backup keeps a copy of unreadable data so a bug never silently destroys a save.
	Backup(b []byte) error
}

// MemStore keeps the save in memory.
type MemStore struct {
	data       []byte
	has        bool
	BackupData []byte
}

func (m *MemStore) Load() ([]byte, error) {
	if !m.has {
		return nil, ErrNoSave
	}
	return append([]byte(nil), m.data...), nil
}

func (m *MemStore) Save(b []byte) error {
	m.data, m.has = append([]byte(nil), b...), true
	return nil
}

func (m *MemStore) Clear() error { m.data, m.has = nil, false; return nil }

func (m *MemStore) Backup(b []byte) error { m.BackupData = append([]byte(nil), b...); return nil }

// FileStore saves to a file, writing atomically via temp file + rename.
type FileStore struct{ Path string }

func (f FileStore) Load() ([]byte, error) {
	b, err := os.ReadFile(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoSave
	}
	return b, err
}

func (f FileStore) Save(b []byte) error {
	dir := filepath.Dir(f.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(f.Path)+".tmp*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after successful rename
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), f.Path)
}

func (f FileStore) Clear() error {
	if err := os.Remove(f.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (f FileStore) Backup(b []byte) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f.Path+".corrupt", b, 0o644)
}
