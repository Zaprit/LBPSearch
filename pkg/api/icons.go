package api

import (
	"bytes"
	"errors"
	hs_img "github.com/HugeSpaceship/HugeSpaceship/pkg/image"
	"image"
	"image/png"

	"github.com/HugeSpaceship/HugeSpaceship/pkg/validation"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/Zaprit/LBPSearch/pkg/utils"
	"io"
	"log/slog"
	"net/http"
)

func IconEndpoint(iconStore storage.ResourceStorageBackend, iconCache storage.IconCacheBackend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		hash := r.PathValue("hash")
		ok, hash := validation.IsHashValid(hash)
		if !ok {
			utils.HttpLog(w, http.StatusBadRequest, "Invalid hash")
			return
		}

		if len(hash) != 40 { // Not a hash
			http.Redirect(w, r, "/static/img/placeholder.png", http.StatusMovedPermanently)
			return
		}

		slog.Info("serving icon", slog.String("hash", hash), slog.Bool("cached", false))

		//hasIcon, err := iconCache.HasIcon(r.Context(), hash)
		//if err != nil {
		//	slog.Error("failed to search for icon", "error", err)
		//	return
		//}
		//if hasIcon {
		//	img, err := iconCache.GetIcon(r.Context(), hash)
		//	if err != nil {
		//		utils.HttpLog(w, http.StatusInternalServerError, "Failed to get icon URL")
		//		return
		//	}
		//	w.Header().Set("Content-Type", "image/png")
		//	w.Header().Set("Cache-Control", "public, max-age=86400")
		//	w.WriteHeader(200)
		//	_, err = io.Copy(w, img)
		//	if err != nil {
		//		slog.Error("failed to write icon image", "error", err)
		//	}
		//	return
		//}

		rawImg, err := iconStore.GetResource(r.Context(), hash)
		if err != nil {
			utils.HttpLog(w, http.StatusInternalServerError, "Failed to get icon")
			slog.Error("Failed to get icon", "error", err)
			return
		}

		img, err := convertImage(rawImg, true)
		if errors.Is(err, hs_img.InvalidMagicNumber) {
			utils.HttpLog(w, http.StatusUnsupportedMediaType, "Not an image")
			return
		} else if err != nil {
			http.Redirect(w, r, "/static/img/placeholder.png", http.StatusTemporaryRedirect)
			slog.Error("failed to load image", slog.Any("error", err), slog.String("hash", hash))
			return
		}
		err = iconCache.PutIcon(r.Context(), hash, img)
		if err != nil {
			slog.Error("failed to save image", slog.Any("error", err), slog.String("hash", hash))
			http.Redirect(w, r, "/static/img/placeholder.png", http.StatusTemporaryRedirect)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.WriteHeader(200)

		img.Seek(0, 0)
		_, err = io.Copy(w, img)
		if err != nil {
			slog.Error("failed to write image", slog.Any("error", err), slog.String("hash", hash))
		}
	}
}

func convertImage(r io.Reader, icon bool) (io.ReadSeeker, error) {
	decompressed, err := hs_img.DecompressImage(r)
	if err != nil {
		return nil, err
	}
	i, _, err := image.Decode(decompressed)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if icon {
		var img = i.(*image.RGBA)
		bounds := img.Bounds()
		size := bounds.Dy()
		xLoc := (bounds.Dx() / 2) - (size / 2)
		rect := image.Rect(xLoc, 0, xLoc+size, size)

		subImg := img.SubImage(rect)

		err := png.Encode(buf, subImg)
		if err != nil {
			return nil, err
		}
	} else {
		err = hs_img.IMGToPNG(decompressed, buf)
		if err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(buf.Bytes()), nil
}
