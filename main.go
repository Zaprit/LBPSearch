package main

import (
	"LBPDumpSearch/model"
	"embed"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//go:embed website
var Files embed.FS

var IndexTemplate = template.Must(template.ParseFS(Files, "website/index.html"))

func IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := IndexTemplate.Execute(w, map[any]any{
			"HasResults": false,
		})
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

func SearchHandler(conn *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		query := fmt.Sprintf("%%%s%%", unescapedquery)

		sort := r.URL.Query().Get("sort")

		where := "(\"npHandle\" ILIKE ?) OR (name ILIKE ?) OR (description ILIKE ?)"
		whereArr := []interface{}{query, query, query}
		if sort == "author" {
			where = "\"npHandle\" ILIKE ?"  // Just override for author, this is janky.
			whereArr = []interface{}{query} // I hate this dearly
		}

		conn.Model(&model.Slot{}).Where(where, whereArr...).Count(&count)

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

			slots[i].Icon = hex.EncodeToString(slot.IconDB)
		}
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
		}

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

		err = IndexTemplate.Execute(w, data)
		if err != nil {
			slog.Error("failed to execute template", slog.Any("error", err))
		}
	}
}

// /mnt/sysdata/dry.db
func main() {
	//db, err := gorm.Open(sqlite.Open("/Users/henry/Downloads/dry.db"))
	//if err != nil {
	//	panic(err)
	//}

	db, err := gorm.Open(postgres.Open("host=localhost user=lbpsearch password=lbpsearch dbname=lbpsearch port=5432 sslmode=disable"))
	if err != nil {
		panic(err)
	}

	//err = db2.AutoMigrate(&model.Slot{})
	//if err != nil {
	//	panic(err)
	//}

	//rows, err := db.Model(&model.Slot{}).Rows()
	//defer rows.Close()
	//
	//c := 1
	//for rows.Next() {
	//	var slot model.Slot
	//	// ScanRows scans a row into a struct
	//	db.ScanRows(rows, &slot)
	//	fmt.Printf("Processing slot #%d \"%s\"\n", c, slot.Name)
	//	c++
	//	db2.Create(&slot)
	//	// Perform operations on each user
	//}

	r := chi.NewRouter()

	r.Get("/", IndexHandler())
	r.Get("/search", SearchHandler(db))
	r.Get("/icon/{hash}", IconHandler())

	err = http.ListenAndServe(":8182", r)
	if err != nil {
		panic(err)
	}
}
