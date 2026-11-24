package wal

import (
	"bytes"
	"testing"
)

func TestSerializeDeserializeFragment(t *testing.T) {
	original := Fragment{
		Type:     FragmentFirst,
		RecordID: 25,
		Data:     []byte("Ovo je dio jednog WAL zapisa"),
	}

	serialized, err := original.SerializeFragment()
	if err != nil {
		t.Fatal(err)
	}

	restored, err := DeserializeFragment(serialized)
	if err != nil {
		t.Fatal(err)
	}

	if restored.Type != original.Type {
		t.Error("fragment type mismatch")
	}

	if restored.RecordID != original.RecordID {
		t.Error("record ID mismatch")
	}

	if restored.Length != uint32(len(original.Data)) {
		t.Error("fragment length mismatch")
	}

	if !bytes.Equal(restored.Data, original.Data) {
		t.Error("fragment data mismatch")
	}

	if restored.CRC != original.CRC {
		t.Error("fragment CRC mismatch")
	}
}

func TestCorruptedFragment(t *testing.T) {
	fragment := Fragment{
		Type:     FragmentFull,
		RecordID: 1,
		Data:     []byte("test"),
	}

	data, err := fragment.SerializeFragment()
	if err != nil {
		t.Fatal(err)
	}

	// Namjerno ostecujemo podatke
	data[17] ^= 0xFF

	_, err = DeserializeFragment(data)

	if err == nil {
		t.Fatal("expected CRC error")
	}
}
