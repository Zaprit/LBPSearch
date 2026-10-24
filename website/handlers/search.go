package handlers

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/Zaprit/LBPSearch/pkg/db"
	"github.com/Zaprit/LBPSearch/website/pages"
	"github.com/a-h/templ"
	"gorm.io/gorm"
)

func SearchHandler(conn *gorm.DB, headerInjection string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var page uint64 = 0
		if pString := r.URL.Query().Get("page"); pString != "" {
			p, err := strconv.ParseUint(pString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Bad Page Number"))
			}
			page = p
		}

		query, err := url.QueryUnescape(r.URL.Query().Get("s"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad Query"))
		}

		if query == "" {
			w.WriteHeader(http.StatusBadRequest)
			panic("TODO: Create Error Page")
			return
		}
		sort := "hearts"
		//sort = r.URL.Query().Get("sort")

		asc := false

		if r.URL.Query().Get("invert") == "on" {
			asc = true
		}

		results, err := db.GetSlots(conn, query, sort, int(page), 30, asc)
		if err != nil {
			panic(err)
		}

		templ.Handler(pages.SearchPage(results, query, headerInjection)).ServeHTTP(w, r)
	}

}
