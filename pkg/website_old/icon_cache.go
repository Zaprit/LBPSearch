package website_old

import (
	"bytes"
	"errors"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/image"
	"github.com/HugeSpaceship/HugeSpaceship/pkg/validation"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/Zaprit/LBPSearch/pkg/utils"
	"io"
	"log/slog"
	"net/http"
)

func IconHandler(iconStore storage.ResourceStorageBackend, iconCache storage.IconCacheBackend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		hash := r.PathValue("hash")
		ok, hash := validation.IsHashValid(hash)
		if !ok {
			utils.HttpLog(w, http.StatusBadRequest, "Invalid hash")
			return
		}

		if len(hash) != 40 { // Not a hash
			http.Redirect(w, r, "/static/placeholder.png", http.StatusMovedPermanently)
			return
		}

		slog.Info("serving icon", slog.String("hash", hash), slog.Bool("cached", false))

		hasIcon, err := iconCache.HasIcon(r.Context(), hash)
		if err != nil {
			slog.Error("failed to search for icon", "error", err)
			return
		}
		if hasIcon {
			url, err := iconCache.GetIconURL(r.Context(), hash)
			if err != nil {
				utils.HttpLog(w, http.StatusInternalServerError, "Failed to get icon URL")
				return
			}
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		rawImg, err := iconStore.GetResource(r.Context(), hash)
		if err != nil {
			utils.HttpLog(w, http.StatusInternalServerError, "Failed to get icon")
			slog.Error("Failed to get icon", "error", err)
			return
		}

		img, err := convertImage(rawImg)
		if errors.Is(err, image.InvalidMagicNumber) {
			utils.HttpLog(w, http.StatusUnsupportedMediaType, "Not an image")
			return
		} else if err != nil {
			http.Redirect(w, r, "/static/placeholder.png", http.StatusTemporaryRedirect)
			slog.Error("failed to load image", slog.Any("error", err), slog.String("hash", hash))
			return
		}

		err = iconCache.PutIcon(r.Context(), hash, img)
		if err != nil {
			slog.Error("failed to save image", slog.Any("error", err), slog.String("hash", hash))
			http.Redirect(w, r, "/static/placeholder.png", http.StatusTemporaryRedirect)
			return
		}

		imgURL, err := iconCache.GetIconURL(r.Context(), hash)
		http.Redirect(w, r, imgURL, http.StatusFound)
	}
}

func convertImage(img io.Reader) (io.Reader, error) {
	decompressed, err := image.DecompressImage(img)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = image.IMGToPNG(decompressed, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
