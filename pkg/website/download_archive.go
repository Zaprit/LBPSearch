package website

import (
	"errors"
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
	"time"
)

var MissingRootLevel = errors.New("rootLevel is missing from archive, rip")

func DownloadArchive(requestID, id, cachePath, dlCommandPath string) (io.ReadSeekCloser, time.Time, string, error) {

	if f, err := os.Open(cachePath + "/levels/" + id + ".zip"); err == nil {
		fmt.Println("returning cached level " + id)
		fi, _ := f.Stat()
		zw, err := zip.NewReader(f, fi.Size())
		if err != nil {
			panic(err)
		}
		return f, fi.ModTime(), strings.ReplaceAll(zw.File[0].Name, "/", "") + ".zip", nil
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
			return nil, time.Time{}, "", MissingRootLevel
		}

		slog.Error("Failed to create backup, please check logs.", "id", id, "err", err, "requestID", requestID)
		return nil, time.Time{}, "", err
	}

	slog.Info("Backup created", "id", string(out))

	f, err := os.OpenFile(path.Join(cachePath, "/levels/"+id+".zip"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, time.Time{}, "", err
	}

	zw := zip.NewWriter(f)

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
		return nil, time.Time{}, "", err
	}
	f.Sync()
	f.Seek(0, 0)

	os.RemoveAll(path.Join(cachePath, "backups", string(out)))

	return f, time.Now(), strings.ReplaceAll(string(out), "/", "") + ".zip", err
}
