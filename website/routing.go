package website

import (
	"embed"
	"mime"
	"net/http"
	"path"

	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/Zaprit/LBPSearch/website/handlers"
	"github.com/Zaprit/LBPSearch/website/pages"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

//go:embed static
var StaticFiles embed.FS

func SetupRouting(cfg *config.Config, router chi.Router, conn *gorm.DB) {
	router.Get("/", templ.Handler(pages.HomePage(cfg.HeaderInjection)).ServeHTTP)

	router.Get("/search", handlers.SearchHandler(conn, cfg.HeaderInjection))

	router.Get("/slot/{id}", handlers.SlotHandler(conn, cfg.HeaderInjection))

	fs := http.FileServerFS(StaticFiles)
	router.Handle("/static/*", http.StripPrefix("/",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=604800, public")
			// Truly, one of the lines of all time
			w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(r.URL.Path)))
			fs.ServeHTTP(w, r)
		}),
	))

}
