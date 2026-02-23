package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Alia5/goandbackup/backup"
	"github.com/Alia5/goandbackup/config"
	"github.com/Alia5/goandbackup/configpaths"
	applog "github.com/Alia5/goandbackup/log"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"
	kongyaml "github.com/alecthomas/kong-yaml"
)

func main() {
	userCfg := findUserConfig(os.Args[1:])
	jsonPaths, yamlPaths, tomlPaths := configpaths.ConfigCandidatePaths(userCfg)

	var cli config.CLI
	ctx := kong.Parse(&cli,
		kong.Name("goandbackup"),
		kong.Description(Description()),
		kong.UsageOnError(),
		kong.Configuration(kong.JSON, jsonPaths...),
		kong.Configuration(kongyaml.Loader, yamlPaths...),
		kong.Configuration(kongtoml.Loader, tomlPaths...),
	)

	_, closeFiles, err := applog.SetupLogger(cli.LogLevel, cli.LogFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to setup logger:", err)
		os.Exit(2)
	}
	defer func() {
		for _, c := range closeFiles {
			_ = c.Close()
		}
	}()

	err = backup.GoAndBackup(&cli)
	if err != nil {
		slog.Error("backup failed", "error", err)
	}
	ctx.FatalIfErrorf(err)
}

func findUserConfig(args []string) string {
	for i, a := range args {
		if strings.HasPrefix(a, "--config=") {
			return a[len("--config="):]
		}
		if a == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return os.Getenv("GOANDBACKUP_CONFIG")
}
