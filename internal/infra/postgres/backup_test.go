package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupManagerBuildsSafeCommands(t *testing.T) {
	root := t.TempDir()
	backup := filepath.Join(root, "daily", "outpipe.dump")
	var name string
	var args []string
	manager, err := NewBackupManager("postgres://database")
	if err != nil {
		t.Fatalf("create backup manager: %v", err)
	}
	manager.runner = func(_ context.Context, command string, values ...string) error {
		name, args = command, values
		return nil
	}
	if err := manager.Backup(context.Background(), backup); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if name != "pg_dump" || len(args) != 4 || args[1] != "--file" || args[2] != backup {
		t.Fatalf("unexpected backup command: %s %v", name, args)
	}
	if err := os.WriteFile(backup, []byte("dump"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := manager.Verify(context.Background(), backup); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if name != "pg_restore" || args[0] != "--list" || args[1] != backup {
		t.Fatalf("unexpected verify command: %s %v", name, args)
	}
}
