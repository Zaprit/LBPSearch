package website

import (
	"fmt"
	"github.com/klauspost/compress/zip"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func DownloadArchive(requestID, id, cachePath, dlCommandPath string) (io.ReadCloser, error) {

	if f, err := os.Open(cachePath + "/levels/" + id + ".zip"); err == nil {
		fmt.Println("returning cached level " + id)
		return f, nil
	}

	logFile, err := os.OpenFile(path.Join(cachePath, "levellogs", id+".log"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	cmd := exec.Command(dlCommandPath, "bkp", id)
	cmd.Dir = cachePath
	cmd.Stderr = logFile

	out, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to create backup, please check logs.", "id", id, "err", err, "requestID", requestID)
		return nil, err
	}

	slog.Info("Backup created", "id", string(out))

	f, err := os.OpenFile(path.Join(cachePath, "/levels/"+id+".zip"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(out))
	zw := zip.NewWriter(f)

	err = filepath.WalkDir(path.Join(cachePath, "backups", string(out)), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			zw.Create(strings.TrimPrefix(p, path.Clean(cachePath+"/backups/")) + "/")
		} else {
			f, err := zw.Create(strings.TrimPrefix(p, path.Clean(cachePath+"/backups/")))
			if err != nil {
				return err
			}
			src, err := os.Open(p)
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
		return nil, err
	}
	f.Sync()
	f.Seek(0, 0)
	return f, err
}
