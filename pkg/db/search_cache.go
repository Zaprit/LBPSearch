package db

import (
	"github.com/Zaprit/LBPSearch/pkg/model"
	"gorm.io/gorm"
)

func GetCount(conn *gorm.DB, query string, authorSort bool) (out int64) {
	conn.Where("(query = ?) AND (author_sort = ?)", query, authorSort).Count(&out)

	if out == 0 {
		where := "(\"npHandle\" ILIKE ?) OR (name ILIKE ?) OR (description ILIKE ?)"
		whereArr := []interface{}{query, query, query}
		if authorSort {
			where = "\"npHandle\" ILIKE ?"  // Just override for author, this is janky.
			whereArr = []interface{}{query} // I hate this dearly
		}

		conn.Model(&model.Slot{}).Where(where, whereArr...).Count(&out)

		if out == 0 {
			out = -1
		}

		conn.Create(model.SearchCache{
			Query:      query,
			Count:      out,
			AuthorSort: authorSort,
		})
	}
	if out == -1 { // There's no results, but we want to cache that.
		out = 0
	}
	return
}
