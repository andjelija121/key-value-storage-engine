package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

const RecordHeaderSize = 29

type Record struct {
	CRC       uint32
	Timestamp uint64
	Tombstone bool
	Key       []byte
	Value     []byte
}

func (r *Record) Serialize() ([]byte, error) {
	keySize := len(r.Key)
	valueSize := len(r.Value)

	totalSize := RecordHeaderSize + keySize + valueSize

	data := make([]byte, totalSize)

	// Timestamp: bajtovi 4-11
	binary.LittleEndian.PutUint64(data[4:12], r.Timestamp) //PutUint64 pretvara broj tipa uint64 u osam bajtova

	// Tombstone: bajt 12
	if r.Tombstone {
		data[12] = 1
	}

	// KeySize: bajtovi 13-20
	binary.LittleEndian.PutUint64(
		data[13:21],
		uint64(keySize),
	)

	// ValueSize: bajtovi 21-28
	binary.LittleEndian.PutUint64(
		data[21:29],
		uint64(valueSize),
	)

	// Key i Value
	copy(data[29:29+keySize], r.Key)
	copy(data[29+keySize:], r.Value)

	// CRC računamo nad svim bajtovima osim CRC polja
	r.CRC = crc32.ChecksumIEEE(data[4:])

	// CRC upisujemo na početak
	binary.LittleEndian.PutUint32(data[0:4], r.CRC)

	return data, nil
}

func Deserialize(data []byte) (*Record, error) {
	// Provjera minimalne duzine zapisa
	if len(data) < RecordHeaderSize {
		return nil, fmt.Errorf("WAL record is too short")
	}

	// Citanje CRC-a
	storedCRC := binary.LittleEndian.Uint32(data[0:4])

	// Provjera integriteta zapisa
	calculatedCRC := crc32.ChecksumIEEE(data[4:])

	if storedCRC != calculatedCRC {
		return nil, fmt.Errorf("WAL record CRC mismatch")
	}

	// Citanje timestampa
	timestamp := binary.LittleEndian.Uint64(data[4:12])

	// Citanje tombstone vrijednosti
	if data[12] != 0 && data[12] != 1 {
		return nil, fmt.Errorf("invalid tombstone value")
	}
	tombstone := data[12] == 1

	// Citanje duzina kljuca i vrijednosti
	keySize := binary.LittleEndian.Uint64(data[13:21])
	valueSize := binary.LittleEndian.Uint64(data[21:29])

	// Provjera da li duzine odgovaraju podacima
	payloadSize := uint64(len(data) - RecordHeaderSize)

	if keySize > payloadSize ||
		valueSize != payloadSize-keySize {
		return nil, fmt.Errorf("invalid WAL record lengths")
	}

	// Izdvajanje kljuca i vrijednosti
	keyEnd := RecordHeaderSize + int(keySize)

	key := make([]byte, int(keySize))
	value := make([]byte, int(valueSize))

	copy(key, data[RecordHeaderSize:keyEnd])
	copy(value, data[keyEnd:])

	// Rekonstrukcija Record objekta
	record := &Record{
		CRC:       storedCRC,
		Timestamp: timestamp,
		Tombstone: tombstone,
		Key:       key,
		Value:     value,
	}

	return record, nil
}
