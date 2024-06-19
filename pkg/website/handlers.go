package website

import (
	"LBPDumpSearch/pkg/config"
	"LBPDumpSearch/pkg/db"
	"LBPDumpSearch/pkg/model"
	"encoding/hex"
	"fmt"
	"github.com/klauspost/compress/gzip"
	"gorm.io/gorm"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IndexHandler serves the root page of the website, it doesn't have to do much as the index is mostly static
func IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		err := IndexTemplate.Execute(w, map[any]any{
			"HasResults": false,
			"Failed":     false,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

// SearchHandler uses the same page as IndexHandler however it does a lot of processing to search for levels
func SearchHandler(conn *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		slots := make([]model.Slot, 0)
		// Total number of matching levels
		var count int64

		var page *uint64 = nil
		if pString := r.URL.Query().Get("page"); pString != "" {
			p, err := strconv.ParseUint(pString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Bad Page Number"))
			}
			page = new(uint64)
			*page = p
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

		// TODO: move DB queries out of web handler, this is bad practice.

		where := "(\"npHandle\" ILIKE ?) OR (name ILIKE ?) OR (description ILIKE ?)"
		whereArr := []interface{}{query, query, query}
		authorSort := false
		if sort == "author" {
			where = "\"npHandle\" ILIKE ?"  // Just override for author, this is janky.
			whereArr = []interface{}{query} // I hate this dearly
			authorSort = true
		}

		count = db.GetCount(conn, query, authorSort)

		q := conn.Limit(50)
		if page != nil {
			q = conn.Offset(int((*page) * 50)).Limit(50)
		}

		q = q.Where(where, whereArr...)

		invert := false

		if r.URL.Query().Get("invert") == "on" {
			invert = true
		}

		switch sort {
		case "name":
			q = q.Order("name ASC")
		case "author":
			q = q.Order("\"npHandle\"")
		case "hearts":
			if invert {
				q = q.Order("\"heartCount\" ASC")
			} else {
				q = q.Order("\"heartCount\" DESC")
			}
		case "published":
			if invert {
				q = q.Order("\"firstPublished\" DESC")
			} else {

			}

		case "updated":
		}

		q.Find(&slots)

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
		// TODO: improve logging, make it quieter where needed and more informative everywhere else.
		slog.Info("New Query",
			slog.String("query", r.URL.Query().Get("s")),
			slog.String("type", r.URL.Query().Get("t")),
			slog.Int64("totalCount", count),
			slog.Int("count", len(slots)),
		)

		data := map[any]any{
			"HasResults":  true,
			"Results":     slots,
			"SearchType":  r.URL.Query().Get("t"),
			"SearchQuery": r.URL.Query().Get("s"),
			"ResultCount": count,
			"Failed":      false,
			"Elapsed":     time.Since(startTime).Round(time.Millisecond).String(),
		}

		// TODO: come up with better way of paginating, as the current count based method is slow, like really slow, like wtf why is this so slow levels of slow.
		if count > 50 {
			data["MaxPage"] = int(float64(count)/50) + 1
			if page != nil {
				data["Page"] = *page
			} else {
				data["Page"] = 0
				page = new(uint64)
			}
			if *page > 0 {
				data["PrevPage"] = *page - 1
			}
			if *page < uint64(int(float64(count)/50)+1) {
				data["NextPage"] = *page + 1
			}
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
			"HasResults":  false,
			"Failed":      false,
			"ShowSponsor": cfg.ShowSponsorMessage,
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

		data := map[any]any{
			"HasResults":  false,
			"Failed":      false,
			"ShowSponsor": cfg.ShowSponsorMessage,
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

func ChangelogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := ChangelogTemplate.Execute(w, template.HTML(Changelog))
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func DownloadArchiveHandler(conn *gorm.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		f, err := DownloadArchive(conn, id, cfg.CachePath, cfg.ArchiveDlCommandPath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		io.Copy(w, f)
		f.Close()
	}
}

func SitemapHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		err := SitemapTemplate.Execute(w, nil)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
func SitemapGZHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		err := SitemapTemplate.Execute(gz, nil)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func SitemapIndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "application/xml")
		err := SitemapIndexTemplate.Execute(w, nil)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
func RobotsTXTHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800, public")
		w.Header().Set("Content-Type", "text/plain")
		err := RobotsTXTTemplate.Execute(w, nil)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}
