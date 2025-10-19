package edit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupManager manages file backups for edit operations.
type BackupManager struct {
	backupDir   string
	maxBackups  int
	retentionDays int
}

// NewBackupManager creates a new backup manager.
func NewBackupManager(workDir string) *BackupManager {
	backupDir := filepath.Join(workDir, ".goai", "backups")
	return &BackupManager{
		backupDir:     backupDir,
		maxBackups:    10,       // Keep max 10 backups per file
		retentionDays: 7,        // Keep backups for 7 days
	}
}

// CreateBackup creates a backup of the specified file.
func (m *BackupManager) CreateBackup(filePath string) (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	originalName := filepath.Base(filePath)
	backupName := fmt.Sprintf("%s.%s.bak", originalName, timestamp)
	backupPath := filepath.Join(m.backupDir, backupName)

	// Copy the file
	if err := m.copyFile(filePath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Clean old backups for this file
	if err := m.cleanOldBackups(originalName); err != nil {
		// Log error but don't fail the backup operation
		fmt.Printf("Warning: failed to clean old backups: %v\n", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a file from backup.
func (m *BackupManager) RestoreBackup(backupPath, targetPath string) error {
	// Verify backup exists
	if _, err := os.Stat(backupPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup file does not exist: %s", backupPath)
		}
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	// Copy backup to target
	if err := m.copyFile(backupPath, targetPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// ListBackups lists all backups for a specific file.
func (m *BackupManager) ListBackups(originalFileName string) ([]BackupInfo, error) {
	// Ensure backup directory exists
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return nil, nil // No backups directory means no backups
	}

	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	prefix := originalFileName + "."

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".bak") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Extract timestamp from filename
			timestamp := extractTimestamp(name, originalFileName)

			backups = append(backups, BackupInfo{
				FileName:     name,
				Path:         filepath.Join(m.backupDir, name),
				OriginalFile: originalFileName,
				Timestamp:    timestamp,
				Size:         info.Size(),
				ModTime:      info.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// GetLatestBackup returns the most recent backup for a file.
func (m *BackupManager) GetLatestBackup(originalFileName string) (*BackupInfo, error) {
	backups, err := m.ListBackups(originalFileName)
	if err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return nil, fmt.Errorf("no backups found for %s", originalFileName)
	}

	return &backups[0], nil
}

// DeleteBackup deletes a specific backup file.
func (m *BackupManager) DeleteBackup(backupPath string) error {
	// Verify it's in the backup directory
	if !strings.HasPrefix(backupPath, m.backupDir) {
		return fmt.Errorf("backup path is not in backup directory")
	}

	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// CleanAllBackups removes all backups older than retention period.
func (m *BackupManager) CleanAllBackups() error {
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return nil // No backup directory, nothing to clean
	}

	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -m.retentionDays)
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove if older than retention period
		if info.ModTime().Before(cutoff) {
			backupPath := filepath.Join(m.backupDir, entry.Name())
			if err := os.Remove(backupPath); err != nil {
				errors = append(errors, fmt.Sprintf("failed to remove %s: %v", entry.Name(), err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// cleanOldBackups removes old backups for a specific file, keeping only maxBackups.
func (m *BackupManager) cleanOldBackups(originalFileName string) error {
	backups, err := m.ListBackups(originalFileName)
	if err != nil {
		return err
	}

	// If we have more than maxBackups, delete the oldest ones
	if len(backups) > m.maxBackups {
		// Backups are already sorted by ModTime (newest first)
		for i := m.maxBackups; i < len(backups); i++ {
			if err := os.Remove(backups[i].Path); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", backups[i].FileName, err)
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst.
func (m *BackupManager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = sourceFile.Close() }()

	// Get source file info
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	// Copy content
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// extractTimestamp extracts the timestamp from a backup filename.
func extractTimestamp(backupName, originalName string) string {
	// Format: originalName.20060102_150405.bak
	prefix := originalName + "."
	suffix := ".bak"

	if strings.HasPrefix(backupName, prefix) && strings.HasSuffix(backupName, suffix) {
		timestamp := backupName[len(prefix) : len(backupName)-len(suffix)]
		// Try to parse and format nicely
		if t, err := time.Parse("20060102_150405", timestamp); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
		return timestamp
	}

	return ""
}

// BackupInfo contains information about a backup.
type BackupInfo struct {
	FileName     string    `json:"filename"`
	Path         string    `json:"path"`
	OriginalFile string    `json:"original_file"`
	Timestamp    string    `json:"timestamp"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
}

// String returns a string representation of BackupInfo.
func (b BackupInfo) String() string {
	return fmt.Sprintf("%s (created: %s, size: %s)",
		b.FileName,
		b.ModTime.Format("2006-01-02 15:04:05"),
		formatFileSize(b.Size))
}

// formatFileSize formats a file size in human-readable format.
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// AutoBackup creates an automatic backup with a descriptive name.
func (m *BackupManager) AutoBackup(filePath string, operation string) (string, error) {
	// Create standard backup first
	backupPath, err := m.CreateBackup(filePath)
	if err != nil {
		return "", err
	}

	// Add operation description to backup metadata
	metadataPath := backupPath + ".meta"
	metadata := fmt.Sprintf("Operation: %s\nTime: %s\nOriginal: %s\n",
		operation,
		time.Now().Format("2006-01-02 15:04:05"),
		filePath)

	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		// Don't fail if metadata can't be written
		fmt.Printf("Warning: failed to write backup metadata: %v\n", err)
	}

	return backupPath, nil
}

// GetBackupMetadata reads the metadata for a backup if it exists.
func (m *BackupManager) GetBackupMetadata(backupPath string) (map[string]string, error) {
	metadataPath := backupPath + ".meta"
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata
		}
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	metadata := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			metadata[parts[0]] = parts[1]
		}
	}

	return metadata, nil
}