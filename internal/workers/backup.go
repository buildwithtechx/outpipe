package workers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type BackupConfig struct {
	Directory     string
	DatabaseURL   string
	PgDumpPath    string
	PgRestorePath string
	Keep          int
}

func (c BackupConfig) Validate() error {

	if strings.TrimSpace(c.Directory) == "" || strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("backup directory and database url are required")
	}

	if c.Keep < 1 {
		return fmt.Errorf("backup keep count must be positive")
	}

	return nil
}

type BackupJob struct {
	config BackupConfig
	runner func(context.Context, string, ...string) ([]byte, error)
	now    func() time.Time
}

func NewBackupJob(config BackupConfig) (*BackupJob, error) {

	if strings.TrimSpace(config.PgDumpPath) == "" {
		config.PgDumpPath = "pg_dump"
	}

	if strings.TrimSpace(config.PgRestorePath) == "" {
		config.PgRestorePath = "pg_restore"
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(config.Directory, 0o750); err != nil {
		return nil, fmt.Errorf("create backup directory: %w", err)
	}

	return &BackupJob{
		config: config,
		runner: runBackupCommand,
		now:    time.Now,
	}, nil
}

func (j *BackupJob) Name() string {
	return "backup"
}

func (j *BackupJob) Run(ctx context.Context) error {
	timestamp := j.now().UTC().Format("20060102T150405Z")
	dumpPath := filepath.Join(j.config.Directory, "outpipe-"+timestamp+".dump")

	if _, err := j.runner(ctx, j.config.PgDumpPath, "--format=custom", "--file="+dumpPath, j.config.DatabaseURL); err != nil {
		return fmt.Errorf("create database backup: %w", err)
	}

	if _, err := j.runner(ctx, j.config.PgRestorePath, "--list", dumpPath); err != nil {
		return fmt.Errorf("verify database backup: %w", err)
	}

	if err := j.prune(ctx); err != nil {
		return err
	}

	return nil
}

func (j *BackupJob) prune(ctx context.Context) error {
	entries, err := os.ReadDir(j.config.Directory)

	if err != nil {
		return fmt.Errorf("list backup directory: %w", err)
	}

	var dumps []string

	for _, entry := range entries {

		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "outpipe-") || !strings.HasSuffix(entry.Name(), ".dump") {
			continue
		}

		dumps = append(dumps, entry.Name())
	}

	sort.Strings(dumps)
	excess := len(dumps) - j.config.Keep

	if excess <= 0 {
		return nil
	}

	for _, name := range dumps[:excess] {

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := os.Remove(filepath.Join(j.config.Directory, name)); err != nil {
			return fmt.Errorf("remove old backup %s: %w", name, err)
		}
	}

	return nil
}

func runBackupCommand(ctx context.Context, executable string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, executable, args...)
	output, err := command.CombinedOutput()

	if err != nil {
		return output, fmt.Errorf("%s: %w: %s", executable, err, strings.TrimSpace(string(output)))
	}

	return output, nil
}
