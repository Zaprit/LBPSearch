package main

import (
	"LBPDumpSearch/pkg/config"
	"LBPDumpSearch/pkg/model"
	"LBPDumpSearch/pkg/website"
	_ "embed"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"os"
	"path"
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

	os.MkdirAll(path.Join(cfg.CachePath, "imgcache"), 0755)
	os.MkdirAll(path.Join(cfg.CachePath, "lvlIcons"), 0755)
	os.MkdirAll(path.Join(cfg.CachePath, "levels"), 0755)

	website.Changelog = changelog
	conn, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DatabaseHost, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabasePort, cfg.DatabaseSSLMode)))
	if err != nil {
		panic(err)
	}
	err = conn.AutoMigrate(&model.SearchCache{}, &model.Slot{})
	if err != nil {
		panic(err)
	}

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

	r.Get("/", website.IndexHandler(cfg))
	r.Get("/search", website.SearchHandler(conn, cfg))
	r.Get("/user/{npHandle}", website.UserHandler(conn, cfg))
	r.Get("/slot/{slotID}", website.SlotHandler(conn, cfg))
	r.Get("/icon/{hash}", website.IconHandler(cfg))
	r.Get("/changelog", website.ChangelogHandler(cfg))
	r.Get("/dl_archive/{id}", website.DownloadArchiveHandler(conn, cfg))
	r.Get("/sitemap_other.xml", website.SitemapHandler(cfg))
	r.Get("/sitemap_other.xml.gz", website.SitemapGZHandler(cfg))
	r.Get("/sitemap.xml", website.SitemapIndexHandler(cfg))
	r.Get("/robots.txt", website.RobotsTXTHandler(cfg))
	r.Get("/static/main.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		http.ServeFileFS(w, r, website.Static, "static/main.css")
	})
	r.Get("/static/JetBrainsMono-Regular.woff2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "font/woff2")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		http.ServeFileFS(w, r, website.Static, "static/JetBrainsMono-Regular.woff2")
	})
	r.Get("/static/refresh.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		http.ServeFileFS(w, r, website.Static, "static/refresh.svg")
	})
	r.Get("/static/refresh.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=604800, public")
		http.ServeFileFS(w, r, website.Static, "static/refresh.png")
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		website.NotFoundTemplate.Execute(w, map[string]interface{}{
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
			"Level":           false,
		})
	})

	err = http.ListenAndServe(":8182", r)
	if err != nil {
		panic(err)
	}
}
