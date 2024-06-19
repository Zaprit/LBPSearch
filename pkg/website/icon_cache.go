package website

import (
	"LBPDumpSearch/pkg/config"
	"bytes"
	"errors"
	"fmt"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/utils"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/utils/file_utils/lbp_image"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/validation"
	"github.com/klauspost/compress/zip"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

func IconHandler(cfg *config.Config) http.HandlerFunc {
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

		if isHourUnsociable() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Hour is currently unsociable, not extracting image to avoid waking hdds."))
			return
		}

		slog.Info("serving icon", slog.String("hash", hash), slog.Bool("cached", false))

		img, err := getImageFromJvyden(hash, cfg.CachePath)
		if errors.Is(err, lbp_image.InvalidMagicNumber) {
			utils.HttpLog(w, http.StatusUnsupportedMediaType, "Not an image")
			return
		} else if err != nil {
			utils.HttpLog(w, http.StatusNotFound, "Image not found")
			slog.Error("failed to load image", slog.Any("error", err), slog.String("hash", hash))
			return
		}

		slog.Info("served icon", slog.String("hash", hash))
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=604800")
		io.Copy(w, img)
	}
}

var imgMutex = new(sync.Mutex)

var InvalidHashError = errors.New("invalid hash")

func getImageFromZip(hash, archivePath, cachePath string) (io.Reader, error) {
	imgMutex.Lock()
	defer imgMutex.Unlock()
	section := hash[0:2]
	section2 := hash[2:4]

	if len(hash) < 40 {
		return nil, InvalidHashError
	}

	var z *zip.Reader
	zipFile, err := os.Open(fmt.Sprintf("%s/dry%s.zip", archivePath, section))
	if err != nil {
		return nil, err
	}
	zfStat, err := zipFile.Stat()
	if err != nil {
		return nil, err
	}

	z, err = zip.NewReader(zipFile, zfStat.Size())
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

	err = os.WriteFile(cachePath+"/imgcache/"+hash, buf.Bytes(), 0644)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func getImageFromJvyden(hash, cachePath string) (io.Reader, error) {
	req, err := http.NewRequest("GET", "https://lbp.littlebigrefresh.com/api/v3/assets/"+hash+"/download", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LBPSearch/1.0 (+https://zaprit.fish; +https://github.com/Zaprit/LBPSearch)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawImg, err := lbp_image.DecompressImage(resp.Body)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = lbp_image.IMGToPNG(rawImg, buf)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(cachePath+"/imgcache/"+hash, buf.Bytes(), 0644)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
