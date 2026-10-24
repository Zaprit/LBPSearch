package db

import (
	"context"
	"errors"

	"github.com/Zaprit/LBPSearch/pkg/model"
	"gorm.io/gorm"
)

func AuthorSort(asc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !asc {
			return db.Order("np_handle ASC")
		} else {
			return db.Order("np_handle DESC")
		}
	}
}

func HeartsSort(asc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if asc {
			return db.Order("heart_count ASC")
		} else {
			return db.Order("heart_count DESC")
		}
	}
}

func PlaysSort(asc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if asc {
			return db.Order("play_count ASC")
		} else {
			return db.Order("play_count DESC")
		}
	}
}

func GetSlots(db *gorm.DB, query, sort string, page, pageSize int, asc bool) ([]model.Slot, error) {
	var tx *gorm.DB
	switch sort {
	case "author":
		tx = db.Scopes(AuthorSort(asc))
	case "hearts":
		tx = db.Scopes(HeartsSort(asc))
	case "plays":
		tx = db.Scopes(PlaysSort(asc))
	default:
		return nil, errors.New("invalid sort type")
	}

	tx.Where("textsearchable_index_col @@ phraseto_tsquery(?)", query)
	tx.Limit(pageSize).Offset((page - 1) * pageSize)
	slots := make([]model.Slot, pageSize)
	tx.Find(&slots)

	for i, s := range slots {
		var author model.User
		db.First(&author, "np_handle = ?", s.NpHandle)
		slots[i].Author = &author
	}

	return slots, tx.Error
}

func GetAdventureLevels(db *gorm.DB) ([]model.Slot, error) {
	slots, err := gorm.G[model.Slot](db).Order("heart_count desc").Where("is_adventure_planet = true").Limit(30).Find(context.Background())
	if err != nil {
		return nil, err
	}
	for i, s := range slots {
		var author model.User
		db.First(&author, "np_handle = ?", s.NpHandle)
		slots[i].Author = &author
	}
	return slots, nil
}
