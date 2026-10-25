package db

import (
	"context"

	"github.com/Zaprit/LBPSearch/pkg/model"
	"gorm.io/gorm"
)

func GetSlot(ctx context.Context, db *gorm.DB, id int32) (model.Slot, error) {
	return gorm.G[model.Slot](db).Where("id = ?", id).Take(ctx)
}
