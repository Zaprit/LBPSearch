package model

import (
	"encoding/hex"
	"time"
)

type User struct {
	NpHandle                  string `gorm:"column:np_handle;primary_key;index"`
	Icon                      []byte `gorm:"column:icon"`
	Lbp1UsedSlots             int64  `gorm:"column:lbp1_used_slots"`
	EntitledSlots             int64  `gorm:"column:entitled_slots"`
	FreeSlots                 int64  `gorm:"column:free_slots"`
	CrossControlUsedSlots     int64  `gorm:"column:cross_control_used_slots"`
	CrossControlEntitledSlots int64  `gorm:"column:cross_control_entitled_slots"`
	CrossControlFreeSlots     int64  `gorm:"column:cross_control_free_slots"`
	LBP2UsedSlots             int64  `gorm:"column:lbp2_used_slots"`
	LBP2EntitledSlots         int64  `gorm:"column:lbp2_entitled_slots"`
	LBP2FreeSlots             int64  `gorm:"column:lbp2_free_slots"`
	LBP3UsedSlots             int64  `gorm:"column:lbp3_used_slots"`
	LBP3EntitledSlots         int64  `gorm:"column:lbp3_entitled_slots"`
	LBP3FreeSlots             int64  `gorm:"column:lbp3_free_slots"`
	HeartCount                int64  `gorm:"column:heart_count"`
	Planets                   []byte `gorm:"column:planets"`
	Yay2                      []byte `gorm:"column:yay2"`
	Boo2                      []byte `gorm:"column:boo2"`
	LocationX                 uint16 `gorm:"column:location_x"`
	LocationY                 uint16 `gorm:"column:location_y"`
	CommentsEnabled           bool   `gorm:"column:comments_enabled"`
	Pins0                     uint32 `gorm:"column:pins0"`
	Pins1                     uint32 `gorm:"column:pins1"`
	Pins2                     uint32 `gorm:"column:pins2"`

	Deleted      bool
	DeletedAt    *time.Time
	DeleteReason string
}

func (u *User) GetIcon() string {
	if u.Icon == nil {
		return ""
	}
	return hex.EncodeToString(u.Icon)
}

func (*User) TableName() string {
	return "users"
}
