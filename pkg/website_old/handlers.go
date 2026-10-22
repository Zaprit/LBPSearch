package website_old

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/Zaprit/LBPSearch/pkg/model"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/google/uuid"
	"github.com/klauspost/compress/gzip"
	"gorm.io/gorm"
)

func UserHandler(conn *gorm.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := model.User{}
		conn.First(&user, "np_handle = ?", r.PathValue("npHandle"))

		data := map[any]any{
			"HasResults":      false,
			"Failed":          false,
			"ShowSponsor":     cfg.ShowSponsorMessage,
			"GlobalURL":       cfg.GlobalURL,
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
		}

		if user.NpHandle != "" {
			data["HasResults"] = true

			slots := make([]model.Slot, 0)
			conn.Where("\"np_handle\" = ?", user.NpHandle).Find(&slots)

			data["Results"] = slots
			data["ResultCount"] = len(slots)

			data["User"] = &user
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "max-age=604800, public")

		err := UserTemplate.Execute(w, data)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func ChangelogHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := ChangelogTemplate.Execute(w, map[string]any{
			"GlobalURL": cfg.GlobalURL,
			"Changelog": template.HTML(Changelog),
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func DownloadArchiveHandler(conn *gorm.DB, cfg *config.Config, backend storage.LevelCacheBackend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		requestID := uuid.New()

		slot := model.Slot{}
		conn.First(&slot, "id = ?", id)
		if slot.ID == 0 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			err := NotFoundTemplate.Execute(w, map[string]any{
				"Level":           true,
				"LevelID":         id,
				"HeaderInjection": template.HTML(cfg.HeaderInjection),
			})
			if err != nil {
				slog.Error("failed to execute template", slog.Any("error", err))
			}
			return
		}

		url, err := DownloadArchive(r.Context(), backend, requestID.String(), id, cfg.CachePath, cfg.ArchiveDlCommandPath)
		if err != nil {
			if errors.Is(err, MissingRootLevel) {
				fmt.Println("Failed to download archive", slog.Any("error", err))
				MissingRootLevelTemplate(w, slot.Name, id, cfg.HeaderInjection, requestID.String())
				slot.Got = 3
				conn.Save(&slot)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			BackupFailTemplate.Execute(w, map[string]any{
				"LevelName":       slot.Name,
				"LevelID":         id,
				"HeaderInjection": template.HTML(cfg.HeaderInjection),
				"RequestID":       requestID.String(),
			})
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}

func SitemapHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		err := SitemapTemplate.Execute(w, map[string]any{
			"GlobalURL": cfg.GlobalURL,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
func SitemapGZHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		err := SitemapTemplate.Execute(gz, map[string]any{
			"GlobalURL": cfg.GlobalURL,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func SitemapIndexHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		err := SitemapIndexTemplate.Execute(w, map[string]any{
			"GlobalURL": cfg.GlobalURL,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
func RobotsTXTHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "text/plain")
		err := RobotsTXTTemplate.Execute(w, map[string]any{
			"GlobalURL": cfg.GlobalURL,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
