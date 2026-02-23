// Package config defines CLI and configuration structures.
package config

// SQLEngine identifies built-in SQL dump engines.
type SQLEngine string

const (
	// Postgres uses pg_dump.
	Postgres SQLEngine = "postgres"
	// MySQL uses mysqldump.
	MySQL SQLEngine = "mysql"
)

// CLI is the root command structure for Kong CLI parsing.
type CLI struct {
	ConfigPath string `help:"Path to configuration file (json|yaml|toml)" name:"config" env:"GOANDBACKUP_CONFIG"`
	BackupName string `help:"Run only one backup job by name" name:"backup" env:"GOANDBACKUP_BACKUP"`
	DryRun     bool   `help:"Validate and print actions without executing" name:"dry-run" env:"GOANDBACKUP_DRY_RUN" default:"false"`
	LogLevel   string `help:"Log level: trace, debug, info, warn, error" default:"info" name:"log-level" env:"GOANDBACKUP_LOG_LEVEL"`
	LogFile    string `help:"Log file path (default: none; logs only to console)" name:"log-file" env:"GOANDBACKUP_LOG_FILE"`

	Backups []Backup `help:"Backup jobs loaded from configuration" name:"backups"`
}

// Backup describes one backup job.
type Backup struct {
	Name       string   `json:"name" yaml:"name" toml:"name"`
	DaysToKeep int      `json:"daysToKeep" yaml:"daysToKeep" toml:"daysToKeep"` // defaults to 30 when zero
	OutDir     string   `json:"outDir" yaml:"outDir" toml:"outDir"`
	SrcFiles   []string `json:"srcFiles" yaml:"srcFiles" toml:"srcFiles"`
	SQLDump    *SQLDump `json:"sqlDump" yaml:"sqlDump" toml:"sqlDump"`
	SQLDumpCmd string   `json:"sqlDumpCmd" yaml:"sqlDumpCmd" toml:"sqlDumpCmd"`
}

// SQLDump configures built-in database dumping.
type SQLDump struct {
	Engine   SQLEngine `json:"engine" yaml:"engine" toml:"engine"`
	Host     string    `json:"host" yaml:"host" toml:"host"`
	Port     int       `json:"port" yaml:"port" toml:"port"`
	User     string    `json:"user" yaml:"user" toml:"user"`
	Password string    `json:"password" yaml:"password" toml:"password"`
	Database string    `json:"database" yaml:"database" toml:"database"`
	Binary   bool      `json:"binary" yaml:"binary" toml:"binary"`
}
