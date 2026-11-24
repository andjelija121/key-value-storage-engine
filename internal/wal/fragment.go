package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

type FragmentType uint8

const (
	FragmentFull FragmentType = iota + 1
	FragmentFirst
	FragmentMiddle
	FragmentLast
)

const (
	FragmentHeaderSize        = 17         //  CRC(4) + Length(4) + Type(1) + RecordID(8)
	CanarySize                = 4          // Canary zauzima 4 bajta
	TrailingCanary     uint32 = 0xDEADBEEF //To je unaprijed definisana vrijednost koju ćemo upisivati na kraj svakog fragmenta
)

type Fragment struct {
	CRC      uint32
	Length   uint32
	Type     FragmentType
	RecordID uint64
	Data     []byte
}

func (f *Fragment) SerializeFragment() ([]byte, error) {
	//provera tipa fragmenta

	if f.Type < FragmentFull || f.Type > FragmentLast {
		return nil, fmt.Errorf("Invalid fragment type: %d", f.Type)

	}

	//provera duzine

	if uint64(len(f.Data)) > uint64(^uint32(0)) {
		return nil, fmt.Errorf("fragment data is too large")
	}
	// Racunamo velicinu podataka
	f.Length = uint32(len(f.Data))

	totalSize := FragmentHeaderSize + len(f.Data) + CanarySize

	data := make([]byte, totalSize)

	// Length: bajtovi 4-7
	binary.LittleEndian.PutUint32(data[4:8], f.Length)

	// Type: bajt 8
	data[8] = byte(f.Type)

	// RecordID: bajtovi 9-16
	binary.LittleEndian.PutUint64(data[9:17], f.RecordID)

	// Podaci fragmenta
	copy(data[17:17+len(f.Data)], f.Data)

	// Canary na kraju fragmenta
	canaryStart := totalSize - CanarySize
	binary.LittleEndian.PutUint32(
		data[canaryStart:],
		TrailingCanary,
	)
	// CRC nad svim bajtovima osim prva 4
	f.CRC = crc32.ChecksumIEEE(data[4:])

	// Upis CRC-a
	binary.LittleEndian.PutUint32(data[0:4], f.CRC)

	return data, nil

}

func DeserializeFragment(data []byte) (*Fragment, error) {
	// Najmanja moguca velicina fragmenta
	minSize := FragmentHeaderSize + CanarySize

	if len(data) < minSize {
		return nil, fmt.Errorf("fragment is too short")
	}

	// Citanje CRC-a
	storedCRC := binary.LittleEndian.Uint32(data[0:4])

	// Provjera CRC-a
	calculatedCRC := crc32.ChecksumIEEE(data[4:])

	if storedCRC != calculatedCRC {
		return nil, fmt.Errorf("fragment CRC mismatch")
	}

	// Citanje duzine
	length := binary.LittleEndian.Uint32(data[4:8])

	// Provjera da li duzina odgovara podacima
	expectedDataSize := len(data) - FragmentHeaderSize - CanarySize

	if uint64(length) != uint64(expectedDataSize) {
		return nil, fmt.Errorf("invalid fragment length")
	}

	// Citanje tipa
	fragmentType := FragmentType(data[8])

	if fragmentType < FragmentFull || fragmentType > FragmentLast {
		return nil, fmt.Errorf("invalid fragment type")
	}

	// Citanje RecordID
	recordID := binary.LittleEndian.Uint64(data[9:17])

	// Provjera canary vrijednosti
	canaryStart := len(data) - CanarySize

	canary := binary.LittleEndian.Uint32(
		data[canaryStart:],
	)

	if canary != TrailingCanary {
		return nil, fmt.Errorf("invalid trailing canary")
	}

	// Izdvajanje podataka
	payload := make([]byte, int(length))

	copy(payload, data[17:canaryStart])

	return &Fragment{
		CRC:      storedCRC,
		Length:   length,
		Type:     fragmentType,
		RecordID: recordID,
		Data:     payload,
	}, nil
}
