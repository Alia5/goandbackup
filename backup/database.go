package backup

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Alia5/goandbackup/config"
)

func goAndBackupDB(b config.Backup, workDir, timestamp string) (string, error) {
	if b.SQLDump == nil && strings.TrimSpace(b.SQLDumpCmd) == "" {
		return "", nil
	}

	if b.SQLDump != nil && b.SQLDump.Binary {
		return "", errBinaryDumpUnsupported
	}

	dumpPath := filepath.Join(workDir, fmt.Sprintf("%s-%s.sql", b.Name, timestamp))
	slog.Debug("creating sql dump", "backup", b.Name, "dumpPath", dumpPath)

	dumpFile, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("open dump file: %w", err)
	}

	var cmd *exec.Cmd
	if strings.TrimSpace(b.SQLDumpCmd) != "" {
		cmd = exec.Command("sh", "-c", b.SQLDumpCmd)
	} else {
		cmd, err = buildDumpCmd(b.SQLDump)
		if err != nil {
			_ = dumpFile.Close()
			return "", err
		}
	}

	cmd.Stdout = dumpFile
	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		_ = dumpFile.Close()
		return "", fmt.Errorf("sql dump command failed: %w: %s", err, stderr.String())
	}

	st, err := dumpFile.Stat()
	if err != nil {
		_ = dumpFile.Close()
		return "", fmt.Errorf("stat dump file: %w", err)
	}
	if err := dumpFile.Close(); err != nil {
		return "", fmt.Errorf("close dump file: %w", err)
	}
	if st.Size() == 0 {
		return "", errSQLDumpEmpty
	}

	return dumpPath, nil
}

func buildDumpCmd(cfg *config.SQLDump) (*exec.Cmd, error) {
	if cfg == nil {
		return nil, errSQLDumpConfigNil
	}

	engine := strings.ToLower(strings.TrimSpace(string(cfg.Engine)))
	if engine == "" {
		switch cfg.Port {
		case 3306:
			engine = string(config.MySQL)
		case 5432:
			engine = string(config.Postgres)
		default:
			engine = string(config.Postgres)
		}
	}

	switch engine {
	case string(config.Postgres), "postgresql", "pg":
		executable := "pg_dump"
		args := []string{}
		if cfg.Host != "" {
			args = append(args, "-h", cfg.Host)
		}
		if cfg.Port > 0 {
			args = append(args, "-p", fmt.Sprintf("%d", cfg.Port))
		}
		if cfg.User != "" {
			args = append(args, "-U", cfg.User)
		}
		if cfg.Database != "" {
			args = append(args, "-d", cfg.Database)
		}
		cmd := exec.Command(executable, args...)
		if cfg.Password != "" {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
		}
		return cmd, nil

	case string(config.MySQL), "mariadb":
		executable := "mysqldump"
		args := []string{}
		if cfg.Host != "" {
			args = append(args, "-h", cfg.Host)
		}
		if cfg.Port > 0 {
			args = append(args, "-P", fmt.Sprintf("%d", cfg.Port))
		}
		if cfg.User != "" {
			args = append(args, "-u", cfg.User)
		}
		if cfg.Database != "" {
			args = append(args, cfg.Database)
		}
		cmd := exec.Command(executable, args...)
		if cfg.Password != "" {
			cmd.Env = append(os.Environ(), "MYSQL_PWD="+cfg.Password)
		}
		return cmd, nil

	default:
		return nil, fmt.Errorf("unsupported sqlDump engine %q (supported: postgres, mysql); use sqlDumpCmd for others", cfg.Engine)
	}
}
