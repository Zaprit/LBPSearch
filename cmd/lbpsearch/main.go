package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/Zaprit/LBPSearch/pkg/db"
	"github.com/Zaprit/LBPSearch/pkg/model"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/Zaprit/LBPSearch/pkg/website_old"
	"github.com/jackc/pgx/v5/pgxpool"
	"html/template"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed CHANGELOG
var changelog string

// /mnt/sysdata/dry.db
func main() {
	//conn, err := gorm.Open(sqlite.Open("/Users/henry/Downloads/dry.conn"))
	//if err != nil {
	//	panic(err)
	//}

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	storageBackend, err := storage.NewS3StorageBackend(cfg)
	if err != nil {
		panic(err)
	}

	os.MkdirAll(path.Join(cfg.CachePath, "levels"), 0755)

	website_old.Changelog = changelog
	conn, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DatabaseHost, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabasePort, cfg.DatabaseSSLMode)))
	if err != nil {
		panic(err)
	}
	err = conn.AutoMigrate(&model.SearchCache{}, &model.Slot{})
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.New(context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DatabaseUser,
			cfg.DatabasePassword,
			cfg.DatabaseHost,
			cfg.DatabasePort,
			cfg.DatabaseName,
			cfg.DatabaseSSLMode),
	)
	if err != nil {
		panic(err)
	}
	queries := db.New(pool)

	//err = db2.AutoMigrate(&model.Slot{})
	//if err != nil {
	//	panic(err)
	//}

	//rows, err := conn.Model(&model.Slot{}).Rows()
	//defer rows.Close()
	//
	//c := 1
	//for rows.Next() {
	//	var slot model.Slot
	//	// ScanRows scans a row into a struct
	//	conn.ScanRows(rows, &slot)
	//	fmt.Printf("Processing slot #%d \"%s\"\n", c, slot.Name)
	//	c++
	//	db2.Create(&slot)
	//	// Perform operations on each user
	//}

	r := chi.NewRouter()

	r.Get("/", website_old.IndexHandler(cfg))
	r.Get("/search", website_old.SearchHandler(queries, cfg))
	r.Get("/user/{npHandle}", website_old.UserHandler(conn, cfg))
	r.Get("/slot/{slotID}", website_old.SlotHandler(conn, cfg))
	r.Get("/icon/{hash}", website_old.IconHandler(storageBackend, storageBackend))
	r.Get("/changelog", website_old.ChangelogHandler(cfg))
	r.Get("/dl_archive/{id}", website_old.DownloadArchiveHandler(conn, cfg, storageBackend))
	r.Get("/sitemap_other.xml", website_old.SitemapHandler(cfg))
	r.Get("/sitemap_other.xml.gz", website_old.SitemapGZHandler(cfg))
	r.Get("/sitemap.xml", website_old.SitemapIndexHandler(cfg))
	r.Get("/robots.txt", website_old.RobotsTXTHandler(cfg))

	fs := http.FileServerFS(website_old.Static)
	r.Handle("/static/*", http.StripPrefix("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=604800, public")
			// Truly, one of the lines of all time
			w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(r.URL.Path)))
			fs.ServeHTTP(w, r)
		}),
	))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		website_old.NotFoundTemplate.Execute(w, map[string]interface{}{
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
			"Level":           false,
		})
	})

	err = http.ListenAndServe(":8182", r)
	if err != nil {
		panic(err)
	}
}
