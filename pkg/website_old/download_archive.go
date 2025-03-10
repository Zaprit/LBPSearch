package website_old

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/klauspost/compress/zip"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var MissingRootLevel = errors.New("rootLevel is missing from archive, rip")

func DownloadArchive(ctx context.Context, backend storage.LevelCacheBackend, requestID, id, cachePath, dlCommandPath string) (string, error) {

	hasLevel, err := backend.HasLevel(ctx, id)
	if err != nil {
		return "", err
	}

	if hasLevel {
		slog.Info("returning cached archive", "requestID", requestID, "id", id)
		url, err := backend.GetLevelURL(ctx, id)
		if err != nil {
			return "", err
		}
		return url, nil
	}

	logFile, err := os.OpenFile(path.Join(cachePath, "levellogs", id+"-"+requestID+".log"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	cmd := exec.Command(dlCommandPath, "bkp", id)
	cmd.Env = append(cmd.Env, "RUST_BACKTRACE=1")
	cmd.Dir = cachePath
	cmd.Stderr = logFile

	out, err := cmd.Output()
	if err != nil {

		logFile.Seek(0, 0)
		log, _ := io.ReadAll(logFile)
		fmt.Println(string(log))
		if strings.Contains(string(log), "rootLevel is missing from the archive, rip") {
			slog.Error("RootLevel is missing from archive, rip", "id", id)
			return "", MissingRootLevel
		}

		slog.Error("Failed to create backup, please check logs.", "id", id, "err", err, "requestID", requestID)
		return "", err
	}

	slog.Info("Backup created", "id", string(out))

	buf := new(bytes.Buffer)

	zw := zip.NewWriter(buf)

	err = filepath.WalkDir(path.Join(cachePath, "backups", string(out)), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		zipFileName := strings.TrimPrefix(p, path.Clean(cachePath+"/backups/")+"/")
		header := zip.FileHeader{
			Name:     zipFileName,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}

		if d.IsDir() {
			header.Name += "/"
			zw.CreateHeader(&header)
		} else {
			f, err := zw.CreateHeader(&header)
			if err != nil {
				return err
			}
			src, err := os.Open(p)
			defer src.Close()
			if err != nil {
				return err
			}
			_, err = io.Copy(f, src)
			if err != nil {
				return err
			}
		}
		return nil
	})

	zw.SetComment("LBP Level Archive from zaprit.fish")
	err = zw.Close()
	if err != nil {
		return "", err
	}

	err = backend.PutLevel(ctx, id, buf)
	if err != nil {
		return "", err
	}

	os.RemoveAll(path.Join(cachePath, "backups", string(out)))

	url, err := backend.GetLevelURL(ctx, id)
	if err != nil {
		return "", err
	}

	return url, nil
}
