package website

import (
	"LBPDumpSearch/pkg/model"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/klauspost/compress/zip"
	"gorm.io/gorm"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Output struct {
	DownloadCount int    `json:"dl_count"`
	FailCount     int    `json:"fail_count"`
	Output        string `json:"output"`
}

func DownloadArchive(conn *gorm.DB, id, cachePath, dlCommandPath string) (io.ReadCloser, error) {

	if f, err := os.Open(cachePath + "/levels/" + id + ".zip"); err == nil {
		fmt.Println("returning cached level " + id)
		return f, nil
	}

	// Get icon for level
	slot := model.Slot{}
	conn.First(&slot, "id = ?", id)
	if slot.ID == 0 {
		return nil, errors.New("level doesn't exist")
	}

	iconHash := hex.EncodeToString(slot.IconDB)

	if _, err := os.Stat(fmt.Sprintf(cachePath + "/lvlIcons/%s.png")); os.IsNotExist(err) && len(iconHash) == 40 {
		iconFile, err := os.OpenFile(fmt.Sprintf(path.Join(cachePath, "/lvlIcons/%s.png"), id), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		icon, err := getImageFromJvyden(hex.EncodeToString(slot.IconDB), cachePath)
		if err != nil {
			return nil, err
		}
		io.Copy(iconFile, icon)
		iconFile.Close()
	}

	cmd := exec.Command(dlCommandPath, "bkp", id)
	cmd.Dir = cachePath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	output := Output{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path.Join(cachePath, "/levels/"+id+".zip"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	err = filepath.WalkDir(path.Join(cachePath, "backups", output.Output), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			zw.Create(strings.TrimPrefix(p, cachePath+"/backups/") + "/")
		} else {
			f, err := zw.Create(strings.TrimPrefix(p, cachePath+"/backups/"))
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

	err = zw.Close()
	f.Close()

	// This is janky
	f, err = os.Open(path.Join(cachePath, "/levels/"+id+".zip"))
	return f, err
}
