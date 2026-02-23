package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Alia5/goandbackup/config"
)

func goAndBackup(b config.Backup, daysToKeep int, now time.Time) error {
	timestamp := now.Format("2006-01-02_150405")
	outDir := expandHome(b.OutDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create outDir: %w", err)
	}

	workDir, err := os.MkdirTemp("", "goandbackup-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	dumpPath, err := goAndBackupDB(b, workDir, timestamp)
	if err != nil {
		return err
	}

	archivePath, err := goAndBackupFiles(outDir, b.Name, timestamp, b.SrcFiles, dumpPath)
	if err != nil {
		return err
	}

	deleted, err := pruneOldBackups(outDir, b.Name, daysToKeep, now)
	if err != nil {
		return err
	}

	slog.Info("backup completed", "backup", b.Name, "archive", archivePath, "pruned", deleted)
	return nil
}

func goAndBackupFiles(outDir, backupName, timestamp string, sources []string, dumpPath string) (string, error) {
	archivePath := filepath.Join(outDir, fmt.Sprintf("%s-%s.tar.gz", backupName, timestamp))
	f, err := os.OpenFile(archivePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open archive file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	if err := goAndAddToArchive(tw, sources); err != nil {
		return "", err
	}
	if err := addDatabaseDumpToArchive(tw, dumpPath); err != nil {
		return "", err
	}

	return archivePath, nil
}

func addDatabaseDumpToArchive(tw *tar.Writer, dumpPath string) error {
	if dumpPath == "" {
		return nil
	}

	fi, err := os.Stat(dumpPath)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(filepath.Join("db", filepath.Base(dumpPath)))
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	df, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(tw, df)
	closeErr := df.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func pruneOldBackups(outDir, backupName string, daysToKeep int, now time.Time) (int, error) {
	if daysToKeep <= 0 {
		return 0, nil
	}

	cutoff := now.Add(-time.Duration(daysToKeep) * 24 * time.Hour)
	pattern := filepath.Join(outDir, backupName+"-*.tar.gz")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0, fmt.Errorf("glob old backups: %w", err)
	}

	deleted := 0
	for _, path := range matches {
		st, err := os.Stat(path)
		if err != nil {
			return deleted, err
		}
		if st.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				return deleted, err
			}
			deleted++
		}
	}

	return deleted, nil
}

func expandHome(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if p == "~" {
		return home
	}
	if len(p) > 1 && p[1] == '/' {
		return filepath.Join(home, p[2:])
	}
	return p
}
