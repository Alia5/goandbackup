package configpaths

import (
	"os"
	"path/filepath"
)

func ConfigCandidatePaths(userPath string) (jsonPaths, yamlPaths, tomlPaths []string) {
	add := func(slice *[]string, p string) {
		if p != "" {
			*slice = append(*slice, p)
		}
	}

	if userPath != "" {
		userPath = expandHome(userPath)
		switch ext := filepath.Ext(userPath); ext {
		case ".json":
			add(&jsonPaths, userPath)
		case ".yaml", ".yml":
			add(&yamlPaths, userPath)
		case ".toml":
			add(&tomlPaths, userPath)
		default:
			add(&jsonPaths, userPath)
		}
	}

	exeDir := executableDir()
	wd, _ := os.Getwd()
	baseDirs := []string{exeDir, wd}
	baseNames := []string{"goandbackup", "config"}

	for _, dir := range baseDirs {
		if dir == "" {
			continue
		}
		for _, base := range baseNames {
			add(&jsonPaths, filepath.Join(dir, base+".json"))
			add(&yamlPaths, filepath.Join(dir, base+".yaml"))
			add(&yamlPaths, filepath.Join(dir, base+".yml"))
			add(&tomlPaths, filepath.Join(dir, base+".toml"))
		}
	}

	return
}

func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
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
