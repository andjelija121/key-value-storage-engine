package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type WALConfig struct {
	Directory string `json:"directory"`

	// Broj blokova u jednom WAL segmentu
	SegmentSizeBlocks int `json:"segment_size_blocks"`

	// Maksimalan broj zapisa u segmentu
	MaxRecordsPerSegment int `json:"max_records_per_segment"`

	SyncOnWrite bool `json:"sync_on_write"`
}

func DefaultWALConfig() WALConfig {
	return WALConfig{
		Directory:            "data/wal",
		SegmentSizeBlocks:    16,
		MaxRecordsPerSegment: 100,
		SyncOnWrite:          true,
	}
}

type Config struct {
	WAL          WALConfig          `json:"wal"`
	BlockManager BlockManagerConfig `json:"block_manager"`
}

func DefaultConfig() Config {
	return Config{
		WAL:          DefaultWALConfig(),
		BlockManager: DefaultBlockManagerConfig(),
	}
}

type BlockManagerConfig struct {
	BlockSize int `json:"block_size"`
}

func DefaultBlockManagerConfig() BlockManagerConfig {
	return BlockManagerConfig{
		BlockSize: 4096,
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := Config{
		WAL: DefaultWALConfig(),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	if cfg.WAL.Directory == "" {
		return Config{}, fmt.Errorf(
			"WAL directory cannot be empty",
		)
	}

	if cfg.WAL.MaxRecordsPerSegment <= 0 {
		return Config{}, fmt.Errorf(
			"max_records_per_segment must be positive",
		)
	}

	return cfg, nil
}
