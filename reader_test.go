package schematic

import (
	"os"
	"testing"
)

func TestSchematic(t *testing.T) {
	filename := "testdata/cylinder.schematic"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Open(\"%s\"): %v", filename, err)
	}
	defer f.Close()
	vol, err := ReadSchematic(f)
	if err != nil {
		t.Fatalf("ReadSchematic: %v", err)
	}
	if vol.XLen() != 128 {
		t.Fatalf("XLen: want 128, but got %d", vol.XLen())
	}
	if !vol.Get(64, 64, 64) {
		t.Fatalf("vol.Get(64,64,64): expected true, but got false")
	}
	if vol.Get(0, 0, 0) {
		t.Fatalf("vol.Get(0,0,0): expected false, but got true")
	}
}
