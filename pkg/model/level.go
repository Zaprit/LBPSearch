package model

type Slot struct {
	ID               uint64 `gorm:"column:id"`
	Name             string `gorm:"column:name"`
	Description      string `gorm:"column:description"`
	NpHandle         string `gorm:"column:npHandle"`
	UploadedIn       string `gorm:"column:publishedIn"`
	Game             uint8  `gorm:"column:game"`
	FirstPublishedDB uint64 `gorm:"column:firstPublished"`
	LastUpdatedDB    uint64 `gorm:"column:lastUpdated"`
	FirstPublished   string `gorm:"-"`
	LastUpdated      string `gorm:"-"`
	HeartCount       uint64 `gorm:"column:heartCount"`
	Background       string `gorm:"column:background"`
	IconDB           []byte `gorm:"column:icon"`
	Icon             string `gorm:"-"`
	RootLevel        []byte `gorm:"column:rootLevel"`
	RootLevelStr     string `gorm:"-"`
	MissingRootLevel bool
}

func (Slot) TableName() string {
	return "slot"
}

type SearchCache struct {
	Query      string `gorm:"column:query"`
	Count      int64  `gorm:"column:count"`
	AuthorSort bool   `gorm:"column:author_sort"`
}
