package backup

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func goAndAddToArchive(tw *tar.Writer, sources []string) error {
	for _, src := range sources {
		src = expandHome(src)
		st, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf("stat srcFile %q: %w", src, err)
		}

		sanitized := filepath.Clean(src)
		sanitized = strings.TrimPrefix(filepath.ToSlash(sanitized), "/")
		sanitized = strings.ReplaceAll(sanitized, ":", "_")
		if sanitized == "." || sanitized == "" {
			sanitized = "root"
		}

		if st.IsDir() {
			prefix := filepath.ToSlash(filepath.Join("sources", sanitized))
			if err := addDirToTar(tw, src, prefix); err != nil {
				return err
			}
			continue
		}

		archivePath := filepath.ToSlash(filepath.Join("sources", sanitized))
		if err := addSingleFileToTar(tw, src, archivePath); err != nil {
			return err
		}
	}
	return nil
}

func addDirToTar(tw *tar.Writer, sourcePath, archivePrefix string) error {
	return filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		rel, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		archivePath := archivePrefix
		if rel != "." {
			archivePath = filepath.ToSlash(filepath.Join(archivePrefix, rel))
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = archivePath
		if info.IsDir() && !strings.HasSuffix(header.Name, "/") {
			header.Name += "/"
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
}

func addSingleFileToTar(tw *tar.Writer, sourceFile, archivePath string) error {
	fi, err := os.Stat(sourceFile)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	header.Name = archivePath
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	f, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
