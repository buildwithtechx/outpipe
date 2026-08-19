package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func fakeBackupRunner(record *[]string, failOn string) func(context.Context, string, ...string) ([]byte, error) {
	return func(ctx context.Context, executable string, args ...string) ([]byte, error) {
		joined := strings.Join(args, " ")
		*record = append(*record, executable+" "+joined)

		if failOn != "" && strings.Contains(joined, failOn) {
			return []byte("simulated failure"), errors.New("exit status 1")
		}

		for _, arg := range args {

			if file, ok := strings.CutPrefix(arg, "--file="); ok {
				_ = os.WriteFile(file, []byte("dump"), 0o600)
			}
		}

		return nil, nil
	}
}

func TestBackupJobRunsDumpAndVerification(t *testing.T) {
	directory := t.TempDir()
	var calls []string
	job, err := NewBackupJob(BackupConfig{Directory: directory, DatabaseURL: "postgres://outpipe:outpipe@localhost:5432/outpipe", Keep: 2})

	if err != nil {
		t.Fatal(err)
	}

	job.runner = fakeBackupRunner(&calls, "")

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("backup run: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected dump and verification commands, got %d calls", len(calls))
	}

	if !strings.Contains(calls[0], "pg_dump") || !strings.Contains(calls[0], "--format=custom") {
		t.Fatalf("unexpected dump command %q", calls[0])
	}

	if !strings.Contains(calls[1], "pg_restore") || !strings.Contains(calls[1], "--list") {
		t.Fatalf("unexpected verification command %q", calls[1])
	}

	files, err := os.ReadDir(directory)

	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || !strings.HasSuffix(files[0].Name(), ".dump") {
		t.Fatalf("expected exactly one backup dump, got %d files", len(files))
	}
}

func TestBackupJobFailsWhenVerificationFails(t *testing.T) {
	directory := t.TempDir()
	var calls []string
	job, err := NewBackupJob(BackupConfig{Directory: directory, DatabaseURL: "postgres://outpipe:outpipe@localhost:5432/outpipe", Keep: 1})

	if err != nil {
		t.Fatal(err)
	}

	job.runner = fakeBackupRunner(&calls, "--list")

	if err := job.Run(context.Background()); err == nil {
		t.Fatal("expected backup verification failure to fail the job")
	}

	if len(calls) != 2 {
		t.Fatalf("expected dump then verification, got %d calls", len(calls))
	}
}

func TestBackupJobPrunesOldDumps(t *testing.T) {
	directory := t.TempDir()

	for _, name := range []string{"outpipe-20260101T000000Z.dump", "outpipe-20260102T000000Z.dump", "outpipe-20260103T000000Z.dump"} {

		if err := os.WriteFile(filepath.Join(directory, name), []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	job, err := NewBackupJob(BackupConfig{Directory: directory, DatabaseURL: "postgres://outpipe:outpipe@localhost:5432/outpipe", Keep: 2})

	if err != nil {
		t.Fatal(err)
	}

	job.runner = fakeBackupRunner(new([]string), "")

	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		t.Fatal(err)
	}

	var names []string

	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 dumps after retention, got %v", names)
	}

	if hasName(names, "outpipe-20260101T000000Z.dump") {
		t.Fatalf("oldest dump should have been pruned: %v", names)
	}
}

func TestBackupJobValidatesConfiguration(t *testing.T) {

	if _, err := NewBackupJob(BackupConfig{DatabaseURL: "postgres://x", Keep: 3}); err == nil {
		t.Fatal("expected directory validation error")
	}

	if _, err := NewBackupJob(BackupConfig{Directory: t.TempDir(), DatabaseURL: "postgres://x"}); err == nil {
		t.Fatal("expected keep count validation error")
	}

	if _, err := NewBackupJob(BackupConfig{Directory: t.TempDir(), DatabaseURL: "postgres://x", Keep: 0}); err == nil {
		t.Fatal("expected non-positive keep validation error")
	}
}

func TestBackupJobNames(t *testing.T) {
	job := &BackupJob{config: BackupConfig{Directory: t.TempDir(), DatabaseURL: "postgres://x", Keep: 1}, runner: fakeBackupRunner(new([]string), ""), now: func() time.Time { return time.Now() }}

	if job.Name() != "backup" {
		t.Fatalf("unexpected job name %q", job.Name())
	}
}

func TestBackupJobReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	job, err := NewBackupJob(BackupConfig{Directory: t.TempDir(), DatabaseURL: "postgres://x", Keep: 1})

	if err != nil {
		t.Fatal(err)
	}

	err = job.Run(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func hasName(names []string, wanted string) bool {

	return slices.Contains(names, wanted)
}
