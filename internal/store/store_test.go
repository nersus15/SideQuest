package store

import (
	"os"
	"path/filepath"
	"testing"
)

type doc struct {
	Name string `json:"name"`
	N    int    `json:"n"`
}

func TestRoundTrip(t *testing.T) {
	f := File{Path: filepath.Join(t.TempDir(), "nested", "state.json")}

	var empty doc
	ok, err := f.Load(&empty)
	if ok || err != nil {
		t.Fatalf("missing file: ok=%v err=%v", ok, err)
	}

	if err := f.Save(doc{Name: "boo", N: 3}); err != nil {
		t.Fatal(err)
	}
	var got doc
	ok, err = f.Load(&got)
	if !ok || err != nil || got.Name != "boo" || got.N != 3 {
		t.Fatalf("round trip failed: %+v ok=%v err=%v", got, ok, err)
	}

	entries, _ := os.ReadDir(filepath.Dir(f.Path))
	if len(entries) != 1 {
		t.Fatalf("temp files left behind: %v", entries)
	}
}

func TestCorruptFileIsMovedAside(t *testing.T) {
	dir := t.TempDir()
	f := File{Path: filepath.Join(dir, "state.json")}
	if err := os.WriteFile(f.Path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	var d doc
	ok, err := f.Load(&d)
	if ok || err == nil {
		t.Fatalf("expected error for corrupt file, ok=%v err=%v", ok, err)
	}
	if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
		t.Fatal("corrupt file should have been moved")
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "state.json.corrupt-*"))
	if len(matches) != 1 {
		t.Fatalf("expected one backup, got %v", matches)
	}
}
