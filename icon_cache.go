package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/utils"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/utils/file_utils/lbp_image"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/validation"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

func IconHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		hash := r.PathValue("hash")
		ok, hash := validation.IsHashValid(hash)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid hash"))
			return
		}

		if len(hash) != 40 { // Not a hash
			utils.HttpLog(w, http.StatusNotFound, "Image not found")
			return
		}

		if _, err := os.Stat("/mnt/sysdata/imgcache/" + hash); err == nil {
			imgFile, err := os.Open("/mnt/sysdata/imgcache/" + hash)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to open image"))
				return
			}
			defer imgFile.Close()
			slog.Info("serving icon", slog.String("hash", hash), slog.Bool("cached", true))
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "max-age=604800")
			io.Copy(w, imgFile)
		}

		slog.Info("serving icon", slog.String("hash", hash), slog.Bool("cached", false))

		img, err := getImageFromZip(hash)
		if errors.Is(err, lbp_image.InvalidMagicNumber) {
			utils.HttpLog(w, http.StatusUnsupportedMediaType, "Not an image")
			return
		} else if err != nil {
			utils.HttpLog(w, http.StatusNotFound, "Image not found")
			slog.Error("failed to load image", slog.Any("error", err))
			return
		}

		slog.Info("served icon", slog.String("hash", hash))
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=604800")
		io.Copy(w, img)
	}
}

var imgMutex = new(sync.Mutex)

func getImageFromZip(hash string) (io.Reader, error) {
	imgMutex.Lock()
	defer imgMutex.Unlock()
	section := hash[0:2]
	section2 := hash[2:4]

	zipFile, err := os.Open(fmt.Sprintf("/mnt/backups/lbparchive/res/dry%s.zip", section))
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()
	zfStat, err := zipFile.Stat()
	if err != nil {
		return nil, err
	}
	z, err := zip.NewReader(zipFile, zfStat.Size())
	if err != nil {
		return nil, err
	}
	imgFile, err := z.Open(fmt.Sprintf("%s/%s/%s", section, section2, hash))
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	decompressed, err := lbp_image.DecompressImage(imgFile)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = lbp_image.IMGToPNG(decompressed, buf)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("/mnt/sysdata/imgcache/"+hash, buf.Bytes(), 0644)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
