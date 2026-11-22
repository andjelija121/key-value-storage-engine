package wal

import (
	"testing"
)

func TestSerialize(t *testing.T) {
	record := Record{
		Timestamp: 123456789,
		Tombstone: false,
		Key:       []byte("ime"),
		Value:     []byte("Ana"),
	}

	data, err := record.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := 29 + 3 + 3

	if len(data) != expectedSize {
		t.Fatalf(
			"expected %d bytes, got %d",
			expectedSize,
			len(data),
		)
	}

	if record.CRC == 0 {
		t.Fatal("CRC was not calculated")
	}
}
