// Package backup implements archiving and retention for configured backup jobs.
package backup

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Alia5/goandbackup/config"
)

const defaultDaysToKeep = 30

var (
	errNoBackupsConfigured   = errors.New("no backups configured")
	errBackupNameEmpty       = errors.New("backup name must not be empty")
	errSQLDumpEmpty          = errors.New("sql dump command produced an empty dump file")
	errSQLDumpConfigNil      = errors.New("sqlDump config is nil")
	errBinaryDumpUnsupported = errors.New("sqlDump.binary must be false (SQL/plain-text dumps only)")
)

// GoAndBackup executes all configured backup jobs (or one selected job).
func GoAndBackup(cli *config.CLI) error {
	if len(cli.Backups) == 0 {
		slog.Error("backup run failed", "error", errNoBackupsConfigured)
		return errNoBackupsConfigured
	}

	slog.Debug("starting backup run", "jobs", len(cli.Backups), "dryRun", cli.DryRun)
	now := time.Now()
	ran := 0

	for _, b := range cli.Backups {
		if cli.BackupName != "" && b.Name != cli.BackupName {
			continue
		}
		ran++

		if strings.TrimSpace(b.Name) == "" {
			slog.Error("invalid backup configuration", "error", errBackupNameEmpty)
			return errBackupNameEmpty
		}
		if strings.TrimSpace(b.OutDir) == "" {
			err := fmt.Errorf("backup %q: outDir must not be empty", b.Name)
			slog.Error("invalid backup configuration", "backup", b.Name, "error", err)
			return err
		}

		hasSources := len(b.SrcFiles) > 0
		hasDump := b.SQLDump != nil || strings.TrimSpace(b.SQLDumpCmd) != ""
		if !hasSources && !hasDump {
			err := fmt.Errorf("backup %q: at least one of srcFiles or sqlDump/sqlDumpCmd is required", b.Name)
			slog.Error("invalid backup configuration", "backup", b.Name, "error", err)
			return err
		}

		daysToKeep := b.DaysToKeep
		if daysToKeep <= 0 {
			daysToKeep = defaultDaysToKeep
		}

		if cli.DryRun {
			slog.Info("dry-run backup", "name", b.Name, "outDir", b.OutDir, "daysToKeep", daysToKeep, "srcFiles", len(b.SrcFiles), "hasDump", hasDump)
			continue
		}

		slog.Info("running backup", "name", b.Name)
		if err := goAndBackup(b, daysToKeep, now); err != nil {
			wrapped := fmt.Errorf("backup %q failed: %w", b.Name, err)
			slog.Error("backup job failed", "backup", b.Name, "error", wrapped)
			return wrapped
		}
	}

	if ran == 0 {
		err := fmt.Errorf("no backup matched selection %q", cli.BackupName)
		slog.Error("backup run failed", "error", err)
		return err
	}

	slog.Info("backup run complete", "executed", ran)
	return nil
}
