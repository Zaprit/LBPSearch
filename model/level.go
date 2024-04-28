package model

type Slot struct {
	ID               uint64 `gorm:"column:id"`
	Name             string `gorm:"column:name index"`
	Description      string `gorm:"column:description"`
	NpHandle         string `gorm:"column:npHandle index"`
	UploadedIn       string `gorm:"column:publishedIn"`
	Game             uint8  `gorm:"column:game"`
	FirstPublishedDB uint64 `gorm:"column:firstPublished"`
	LastUpdatedDB    uint64 `gorm:"column:lastUpdated"`
	FirstPublished   string `gorm:"-"`
	LastUpdated      string `gorm:"-"`
	HeartCount       uint64 `gorm:"column:heartCount"`
	Background       string `gorm:"column:background"`
}

func (Slot) TableName() string {
	return "slot"
}
