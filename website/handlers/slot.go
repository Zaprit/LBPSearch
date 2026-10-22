package handlers

import (
	"github.com/Zaprit/LBPSearch/pkg/model"
	"github.com/Zaprit/LBPSearch/website/pages"
	"github.com/a-h/templ"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func SlotHandler(conn *gorm.DB, headerInjection string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad ID"))
			return
		}
		numId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad ID"))
		}

		var slot model.Slot
		conn.Find(&slot, "id = ?", numId)
		if slot.ID == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
			return
		}

		templ.Handler(pages.SlotPage(slot, headerInjection)).ServeHTTP(w, r)

	}
}
