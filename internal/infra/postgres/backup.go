package postgres

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type CommandRunner func(context.Context, string, ...string) error

type BackupManager struct {
	dsn    string
	runner CommandRunner
}

func NewBackupManager(dsn string) (*BackupManager, error) {

	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}

	return &BackupManager{dsn: dsn, runner: runCommand}, nil
}

func (m *BackupManager) Backup(ctx context.Context, destination string) error {

	if m == nil || m.dsn == "" || destination == "" {
		return fmt.Errorf("backup manager and destination are required")
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}

	if err := m.runner(ctx, "pg_dump", "--format=custom", "--file", destination, m.dsn); err != nil {
		return fmt.Errorf("create postgres backup: %w", err)
	}

	return nil
}

func (m *BackupManager) Verify(ctx context.Context, backup string) error {

	if m == nil || backup == "" {
		return fmt.Errorf("backup manager and backup path are required")
	}

	if _, err := os.Stat(backup); err != nil {
		return fmt.Errorf("stat postgres backup: %w", err)
	}

	if err := m.runner(ctx, "pg_restore", "--list", backup); err != nil {
		return fmt.Errorf("verify postgres backup: %w", err)
	}

	return nil
}

func (m *BackupManager) Restore(ctx context.Context, backup string, targetDSN string) error {

	if m == nil || backup == "" || targetDSN == "" {
		return fmt.Errorf("backup, target dsn, and backup manager are required")
	}

	if err := m.runner(ctx, "pg_restore", "--clean", "--if-exists", "--dbname", targetDSN, backup); err != nil {
		return fmt.Errorf("restore postgres backup: %w", err)
	}

	return nil
}

func runCommand(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, string(output))
	}

	return nil
}
