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

// ImportSlot is the raw slot from the dry.db archive file
type ImportSlot struct {
	ID                  uint32
	NPHandle            string
	LocationX           uint16
	LocationY           uint16
	Game                uint8
	Name                *string
	Description         *string
	RootLevel           []byte
	Icon                []byte
	InitiallyLocked     bool
	IsSubLevel          bool
	IsLBP1Only          bool
	Background          uint32
	Thumbnail           []byte
	Shareable           bool
	AuthorLabels        []byte
	LevelType           *string
	MinPlayers          uint8
	MaxPlayers          uint8
	IsAdventurePlanet   bool
	PS4Only             bool
	HeartCount          uint32
	ThumbsUp            uint32
	ThumbsDown          uint32
	AverageRating       float64
	MMPick              bool
	ReviewCount         uint32
	CommentCount        uint32
	PhotoCount          uint32
	AuthorPhotoCount    uint32
	Tags                []byte
	Labels              []byte
	FirstPublished      uint64
	LastUpdated         uint64
	Genre               *string
	CommentsEnabled     bool
	ReviewsEnabled      bool
	PublishedIn         *string
	PlayCount           uint32
	CompletionCount     uint32
	LBP1PlayCount       uint32
	LBP1CompletionCount uint32
	LBP1UniquePlayCount uint32
	LBP2PlayCount       uint32
	LBP2CompletionCount uint32
	UniquePlayCount     uint32
	LBP3PlayCount       uint32
	LBP3CompletionCount uint32
	LBP3UniquePlayCount uint32
	Got                 bool
}
