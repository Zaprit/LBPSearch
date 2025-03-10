package website_old

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/Zaprit/LBPSearch/pkg/db"
	"github.com/Zaprit/LBPSearch/pkg/model"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/google/uuid"
	"github.com/klauspost/compress/gzip"
	"gorm.io/gorm"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IndexHandler serves the root page of the website, it doesn't have to do much as the index is mostly static
func IndexHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		err := IndexTemplate.Execute(w, map[any]any{
			"HasResults":      false,
			"Failed":          false,
			"GlobalURL":       cfg.GlobalURL,
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

// SearchHandler uses the same page as IndexHandler however it does a lot of processing to search for levels
func SearchHandler(conn *db.Queries, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		var page uint64 = 0
		if pString := r.URL.Query().Get("page"); pString != "" {
			p, err := strconv.ParseUint(pString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Bad Page Number"))
			}
			page = p
		}

		unescapedquery, err := url.QueryUnescape(r.URL.Query().Get("s"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad Query"))
		}

		if unescapedquery == "" {
			w.WriteHeader(http.StatusBadRequest)
			err = IndexTemplate.Execute(w, map[string]interface{}{
				"Error":  "Query is too short",
				"Failed": true,
			})
			return
		}
		query := fmt.Sprintf("%s%%", unescapedquery)

		sort := r.URL.Query().Get("sort")

		direction := "DESC"

		if r.URL.Query().Get("invert") == "on" {
			direction = "ASC"
		}

		// TODO: move DB queries out of web handler, this is bad practice.

		var results []db.Slot

		results, err = conn.GetSlotsSort(r.Context(), db.GetSlotsSortParams{
			SearchQuery:     query,
			SearchColumn:    sort,
			SearchDirection: direction,
			SearchOffset:    int32(page * 50),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed to execute template", slog.Any("error", err))
		}

		slots := make([]model.Slot, len(results))
		for i, result := range results {
			slots[i] = model.Slot{
				ID:               uint64(result.ID),
				Name:             result.Name,
				Description:      result.Description,
				NpHandle:         result.NpHandle,
				HeartCount:       uint64(result.HeartCount),
				Background:       result.Background,
				RootLevel:        result.RootLevel,
				RootLevelStr:     hex.EncodeToString(result.RootLevel),
				MissingRootLevel: result.MissingRootLevel,
			}

			slots[i].FirstPublished = time.UnixMilli(result.FirstPublished).Format(time.DateTime)
			slots[i].LastUpdated = time.UnixMilli(result.LastUpdated).Format(time.DateTime)

			if result.Name == "" {
				slots[i].Name = "Unnamed Level"
			}

			// The data in the backup was inconsistent with this, some levels had the game field filled out, some had the UploadedIn filled out
			if result.PublishedIn == "" {
				switch result.Game {
				case 0:
					slots[i].UploadedIn = "LittleBigPlanet"
				case 1:
					slots[i].UploadedIn = "LittleBigPlanet 2"
				case 2:
					slots[i].UploadedIn = "LittleBigPlanet 3"

				}
			} else {
				switch result.PublishedIn {
				case "lbp2":
					slots[i].UploadedIn = "LittleBigPlanet 2"
				case "lbp3ps3":
					slots[i].UploadedIn = "LittleBigPlanet 3 PS3"
				case "lbp3ps4":
					slots[i].UploadedIn = "LittleBigPlanet 3 PS4/PS5"
				}
			}

			slots[i].Icon = hex.EncodeToString(result.Icon) // Icon is conveniently stored as a bytea (BLOB in sqlite) in the archive, this sucks for sending the thing out but what can you do.
		}
		// TODO: improve logging, make it quieter where needed and more informative everywhere else.
		slog.Info("New Query",
			slog.String("query", r.URL.Query().Get("s")),
			slog.String("type", r.URL.Query().Get("t")),
			slog.Int("count", len(slots)),
		)

		data := map[any]any{
			"HasResults":      true,
			"Results":         slots,
			"SearchType":      r.URL.Query().Get("sort"),
			"SearchQuery":     r.URL.Query().Get("s"),
			"Failed":          false,
			"Elapsed":         time.Since(startTime).Round(time.Millisecond).String(),
			"GlobalURL":       cfg.GlobalURL,
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
			"Page":            page + 1,
		}

		// TODO: come up with better way of paginating, as the current count based method is slow, like really slow, like wtf why is this so slow levels of slow.
		if len(results) == 50 {
			data["Page"] = page
			data["NextPage"] = page + 1
		}
		if page > 0 {
			data["PrevPage"] = page - 1
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "max-age=604800, public")

		err = IndexTemplate.Execute(w, data)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

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
			conn.Where("\"npHandle\" = ?", user.NpHandle).Find(&slots)

			for i, slot := range slots {
				slots[i].FirstPublished = time.UnixMilli(int64(slot.FirstPublishedDB)).Format(time.DateTime)
				slots[i].LastUpdated = time.UnixMilli(int64(slot.LastUpdatedDB)).Format(time.DateTime)

				if slot.Name == "" {
					slots[i].Name = "Unnamed Level"
				}

				// The data in the backup was inconsistent with this, some levels had the game field filled out, some had the UploadedIn filled out
				if slot.UploadedIn == "" {
					switch slot.Game {
					case 0:
						slots[i].UploadedIn = "LittleBigPlanet"
					case 1:
						slots[i].UploadedIn = "LittleBigPlanet 2"
					case 2:
						slots[i].UploadedIn = "LittleBigPlanet 3"

					}
				} else {
					switch slot.UploadedIn {
					case "lbp2":
						slots[i].UploadedIn = "LittleBigPlanet 2"
					case "lbp3ps3":
						slots[i].UploadedIn = "LittleBigPlanet 3 PS3"
					case "lbp3ps4":
						slots[i].UploadedIn = "LittleBigPlanet 3 PS4/PS5"
					}
				}

				slots[i].Icon = hex.EncodeToString(slot.IconDB) // Icon is conveniently stored as a bytea (BLOB in sqlite) in the archive, this sucks for sending the thing out but what can you do.
			}

			data["Results"] = slots
			data["ResultCount"] = len(slots)

			user.Icon = hex.EncodeToString(user.IconDB)
			data["User"] = user
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

func SlotHandler(conn *gorm.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slot := model.Slot{}
		conn.First(&slot, "id = ?", r.PathValue("slotID"))

		data := map[string]any{
			"HasResults":      false,
			"Failed":          false,
			"ShowSponsor":     cfg.ShowSponsorMessage,
			"GlobalURL":       cfg.GlobalURL,
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
		}

		if slot.ID != 0 {
			data["HasResults"] = true
			if slot.Name == "" {
				slot.Name = "Unnamed Level"
			}

			if slot.UploadedIn == "" {
				switch slot.Game {
				case 0:
					slot.UploadedIn = "LittleBigPlanet"
				case 1:
					slot.UploadedIn = "LittleBigPlanet 2"
				case 2:
					slot.UploadedIn = "LittleBigPlanet 3"

				}
			} else {
				switch slot.UploadedIn {
				case "lbp2":
					slot.UploadedIn = "LittleBigPlanet 2"
				case "lbp3ps3":
					slot.UploadedIn = "LittleBigPlanet 3 PS3"
				case "lbp3ps4":
					slot.UploadedIn = "LittleBigPlanet 3 PS4/PS5"
				}
			}

			user := model.User{}
			conn.First(&user, "np_handle = ?", slot.NpHandle)
			data["UserIcon"] = hex.EncodeToString(user.IconDB)

			slot.FirstPublished = time.UnixMilli(int64(slot.FirstPublishedDB)).Format(time.DateTime)
			slot.LastUpdated = time.UnixMilli(int64(slot.LastUpdatedDB)).Format(time.DateTime)

			slot.Icon = hex.EncodeToString(slot.IconDB)

			slot.RootLevelStr = hex.EncodeToString(slot.RootLevel)
			// TODO: this is hateful, and wasteful, and sucks
			data["DownloadLink"] = "https://archive.org/download/dry23r" + string(slot.RootLevelStr[0]) + "/dry" + slot.RootLevelStr[0:2] + ".zip/" + slot.RootLevelStr[0:2] + "%2F" + slot.RootLevelStr[2:4] + "%2F" + slot.RootLevelStr
			data["Result"] = slot
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "max-age=604800, public")

		err := SlotTemplate.Execute(w, data)
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
				slot.MissingRootLevel = true
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
