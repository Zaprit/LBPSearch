package handlers

import (
	"fmt"
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
		var page uint64 = 1

		if pString := r.URL.Query().Get("page"); pString != "" {
			p, err := strconv.ParseUint(pString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Bad Page Number"))
			}
			page = p
		}
		var maxPage uint64 = page
		if pString := r.URL.Query().Get("maxpage"); pString != "" {
			p, err := strconv.ParseUint(pString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Bad Max Page Number"))
			}
			if p >= page {
				maxPage = p
			}
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
		sort := r.URL.Query().Get("sort")

		games := r.URL.Query()["game"]
		lbp1 := false
		lbp2 := false
		lbp3 := false
		for _, game := range games {
			switch game {
			case "lbp1":
				lbp1 = true
			case "lbp2":
				lbp2 = true
			case "lbp3":
				lbp3 = true
			}
		}

		asc := false

		if r.URL.Query().Get("invert") == "on" {
			asc = true
		}

		results, err := db.GetSlots(conn, query, sort, int(page), 30, asc)
		if err != nil {
			panic(err)
		}

		urlQuery := r.URL.Query()
		urlQuery.Set("maxpage", fmt.Sprintf("%d", maxPage))
		r.URL.RawQuery = urlQuery.Encode()

		prevPageURL := r.URL.Clone()
		urlQuery = prevPageURL.Query()
		urlQuery.Set("page", fmt.Sprintf("%d", page-1))
		prevPageURL.RawQuery = urlQuery.Encode()

		nextPageURL := r.URL.Clone()
		urlQuery2 := nextPageURL.Query()
		urlQuery2.Set("page", fmt.Sprintf("%d", page+1))
		nextPageURL.RawQuery = urlQuery2.Encode()

		templ.Handler(pages.SearchPage(results, query, headerInjection, r.URL.Query().Has("filters-shown"), sort, asc, lbp1, lbp2, lbp3, page, maxPage, prevPageURL, nextPageURL)).ServeHTTP(w, r)
	}

}
