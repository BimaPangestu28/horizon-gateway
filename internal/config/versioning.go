package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type VersionedConfig struct {
	Version    string      `yaml:"version" json:"version"`
	ConfigHash string      `yaml:"config_hash" json:"config_hash"`
	ModifiedAt time.Time   `yaml:"modified_at" json:"modified_at"`
	Config     *Config     `yaml:"config" json:"config"`
	Metadata   interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type ConfigHistory struct {
	BackupID  string    `json:"backup_id"`
	Version   string    `json:"version"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

func SaveConfig(cfg *VersionedConfig, path string) error {
	if cfg.ConfigHash == "" {
		hash, err := calculateConfigHash(cfg.Config)
		if err != nil {
			return fmt.Errorf("calculating config hash: %w", err)
		}
		cfg.ConfigHash = hash
	}

	if cfg.ModifiedAt.IsZero() {
		cfg.ModifiedAt = time.Now()
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func LoadVersionedConfig(path string) (*VersionedConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var versionedConfig VersionedConfig
	err = yaml.Unmarshal(data, &versionedConfig)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &versionedConfig, nil
}

func calculateConfigHash(cfg *Config) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshaling config for hash: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func BackupConfig(cfg *VersionedConfig, backupsDir string) (string, error) {
	err := os.MkdirAll(backupsDir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating backups directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupID := fmt.Sprintf("backup_%s", timestamp)
	filename := fmt.Sprintf("%s_v%s.yaml", backupID, cfg.Version)
	path := filepath.Join(backupsDir, filename)

	err = SaveConfig(cfg, path)
	if err != nil {
		return "", fmt.Errorf("saving backup config: %w", err)
	}

	err = updateBackupIndex(backupsDir, ConfigHistory{
		BackupID:  backupID,
		Version:   cfg.Version,
		Hash:      cfg.ConfigHash,
		Timestamp: time.Now(),
		Path:      path,
	})
	if err != nil {
		return "", fmt.Errorf("updating backup index: %w", err)
	}

	return backupID, nil
}

func GetConfigBackups(backupsDir string) ([]ConfigHistory, error) {
	indexPath := filepath.Join(backupsDir, "backup_index.json")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return []ConfigHistory{}, nil
	}

	data, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("reading backup index: %w", err)
	}

	var backups []ConfigHistory
	err = json.Unmarshal(data, &backups)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling backup index: %w", err)
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

func RestoreConfigBackup(backupID string, backupsDir string) (*VersionedConfig, error) {
	backups, err := GetConfigBackups(backupsDir)
	if err != nil {
		return nil, fmt.Errorf("getting config backups: %w", err)
	}

	var backup ConfigHistory
	found := false

	for _, b := range backups {
		if b.BackupID == backupID {
			backup = b
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("backup with ID %s not found", backupID)
	}

	return LoadVersionedConfig(backup.Path)
}

func updateBackupIndex(backupsDir string, newBackup ConfigHistory) error {
	indexPath := filepath.Join(backupsDir, "backup_index.json")

	var backups []ConfigHistory

	if _, err := os.Stat(indexPath); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(indexPath)
		if err != nil {
			return fmt.Errorf("reading backup index: %w", err)
		}

		err = json.Unmarshal(data, &backups)
		if err != nil {
			return fmt.Errorf("unmarshaling backup index: %w", err)
		}
	}

	backups = append(backups, newBackup)

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	maxBackups := 20
	if len(backups) > maxBackups {
		backups = backups[:maxBackups]
	}

	data, err := json.MarshalIndent(backups, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling backup index: %w", err)
	}

	err = ioutil.WriteFile(indexPath, data, 0644)
	if err != nil {
		return fmt.Errorf("writing backup index: %w", err)
	}

	return nil
}
