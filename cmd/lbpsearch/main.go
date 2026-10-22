package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Zaprit/LBPSearch/pkg/api"
	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/Zaprit/LBPSearch/pkg/storage"
	"github.com/Zaprit/LBPSearch/pkg/website_old"
	"github.com/Zaprit/LBPSearch/website"
)

//go:embed CHANGELOG
var changelog string

// /mnt/sysdata/dry.db
func main() {
	slog.Info("Starting lbpsearch")

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	storageBackend, err := storage.NewS3StorageBackend(cfg)
	if err != nil {
		panic(err)
	}

	slog.Info("Connecting to the database", "name", cfg.DatabaseUser)
	website_old.Changelog = changelog
	conn, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s dbname=%s user=%s password=%s port=%s sslmode=%s",
		cfg.DatabaseHost, cfg.DatabaseName, cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabasePort, cfg.DatabaseSSLMode)))
	if err != nil {
		panic(err)
	}

	//err = conn.AutoMigrate(&model.Slot{}, &model.User{})
	//if err != nil {
	//	panic(err)
	//}

	slog.Info("Creating indexes")
	//conn.Raw("ALTER TABLE slots ADD COLUMN IF NOT EXISTS textsearchable_lang text NOT NULL DEFAULT 'english';")
	//conn.Raw("ALTER TABLE slots ADD COLUMN IF NOT EXISTS textsearchable_index_col tsvector GENERATED ALWAYS AS (to_tsvector(slots.textsearchable_lang, coalesce(slots.name, '') || ' ' || coalesce(slots.np_handle, '') || ' ' || coalesce(slots.description, ''))) STORED;")
	//conn.Raw("CREATE INDEX slots_textsearch_idx ON slots USING GIN (textsearchable_index_col);")
	//conn.Raw("CREATE INDEX slot_description_gin_trgm_idx ON slots USING GIN (description gin_trgm_ops);")
	//conn.Raw("CREATE INDEX slot_name_gin_trgm_idx ON slots USING GIN (name gin_trgm_ops);")
	//conn.Raw("CREATE INDEX slot_nphandle_gin_trgm_idx ON slots USING GIN (np_handle gin_trgm_ops);")
	//conn.Raw("CREATE INDEX binley_mega_index ON slots USING GIN (np_handle gin_trgm_ops);")
	conn.Debug()
	slog.Info("Connected to the database")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Compress(5))

	r.Get("/api/resource/{resource}", api.ResourcesEndpoint())

	website.SetupRouting(cfg, r, conn)

	//r.Get("/", website_old.IndexHandler(cfg))
	//r.Get("/search", website_old.SearchHandler(conn, cfg))
	r.Get("/user/{npHandle}", website_old.UserHandler(conn, cfg))
	//r.Get("/slot/{slotID}", website_old.SlotHandler(conn, cfg))
	r.Get("/icons/{hash}", api.IconEndpoint(storageBackend, storageBackend))
	r.Get("/changelog", website_old.ChangelogHandler(cfg))
	r.Get("/dl_archive/{id}", website_old.DownloadArchiveHandler(conn, cfg, storageBackend))
	r.Get("/sitemap_other.xml", website_old.SitemapHandler(cfg))
	r.Get("/sitemap_other.xml.gz", website_old.SitemapGZHandler(cfg))
	r.Get("/sitemap.xml", website_old.SitemapIndexHandler(cfg))
	r.Get("/robots.txt", website_old.RobotsTXTHandler(cfg))

	//fs := http.FileServerFS(website_old.Static)
	//r.Handle("/static/*", http.StripPrefix("/",
	//	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		w.Header().Set("Cache-Control", "max-age=604800, public")
	//		// Truly, one of the lines of all time
	//		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(r.URL.Path)))
	//		fs.ServeHTTP(w, r)
	//	}),
	//))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := website_old.NotFoundTemplate.Execute(w, map[string]interface{}{
			"HeaderInjection": template.HTML(cfg.HeaderInjection),
			"Level":           false,
		})
		if err != nil {
			panic(err)
		}
	})

	slog.Info("Started server", "port", 8182)
	err = http.ListenAndServe(":8182", r) //todo: make listen port configurable
	if err != nil {
		panic(err)
	}
}
