package model

type ImportSlot struct {
	ID                  uint32  `gorm:"column:id"`
	NpHandle            string  `gorm:"column:npHandle"`
	LocationX           int64   `gorm:"column:locationX"`
	LocationY           int64   `gorm:"column:locationY"`
	Game                uint8   `gorm:"column:game"`
	Name                string  `gorm:"column:name"`
	Description         string  `gorm:"column:description"`
	RootLevel           []byte  `gorm:"column:rootLevel"`
	Icon                []byte  `gorm:"column:icon"`
	InitiallyLocked     bool    `gorm:"column:initiallyLocked"`
	IsSubLevel          bool    `gorm:"column:isSubLevel"`
	IsLBP1Only          bool    `gorm:"column:isLBP1Only"`
	Background          int64   `gorm:"column:background"`
	Thumbnail           []byte  `gorm:"column:thumbnail"`
	Shareable           bool    `gorm:"column:shareable"`
	AuthorLabels        []byte  `gorm:"column:authorLabels"`
	Leveltype           string  `gorm:"column:leveltype"`
	MinPlayers          uint8   `gorm:"column:minPlayers"`
	MaxPlayers          uint8   `gorm:"column:maxPlayers"`
	IsAdventurePlanet   bool    `gorm:"column:isAdventurePlanet"`
	PS4Only             bool    `gorm:"column:ps4Only"`
	HeartCount          int64   `gorm:"column:heartCount"`
	ThumbsUp            int64   `gorm:"column:thumbsup"`
	ThumbsDown          int64   `gorm:"column:thumbsdown"`
	AverageRating       float64 `gorm:"column:averageRating"`
	MMPick              bool    `gorm:"column:mmpick"`
	ReviewCount         int64   `gorm:"column:reviewCount"`
	CommentCount        int64   `gorm:"column:commentCount"`
	PhotoCount          int64   `gorm:"column:photoCount"`
	AuthorPhotoCount    int64   `gorm:"column:authorPhotoCount"`
	Tags                []byte  `gorm:"column:tags"`
	Labels              []byte  `gorm:"column:labels"`
	FirstPublished      uint64  `gorm:"column:firstPublished"`
	LastUpdated         uint64  `gorm:"column:lastUpdated"`
	Genre               string  `gorm:"column:genre"`
	CommentsEnabled     bool    `gorm:"column:commentsEnabled"`
	ReviewsEnabled      bool    `gorm:"column:reviewsEnabled"`
	PublishedIn         string  `gorm:"column:publishedIn"`
	PlayCount           int64   `gorm:"column:playCount"`
	CompletionCount     int64   `gorm:"column:completionCount"`
	LBP1PlayCount       int64   `gorm:"column:lbp1PlayCount"`
	LBP1CompletionCount int64   `gorm:"column:lbp1CompletionCount"`
	LBP1UniquePlayCount int64   `gorm:"column:lbp1UniquePlayCount"`
	LBP2PlayCount       int64   `gorm:"column:lbp2PlayCount"`
	LBP2CompletionCount int64   `gorm:"column:lbp2CompletionCount"`
	UniquePlayCount     int64   `gorm:"column:uniquePlayCount"`
	LBP3PlayCount       int64   `gorm:"column:lbp3PlayCount"`
	LBP3CompletionCount int64   `gorm:"column:lbp3CompletionCount"`
	LBP3UniquePlayCount int64   `gorm:"column:lbp3UniquePlayCount"`
	Got                 uint8   `gorm:"column:got"`
}

func (*ImportSlot) TableName() string {
	return "slot"
}

type ImportUser struct {
	NpHandle                  string `gorm:"column:npHandle;primaryKey"`
	Icon                      []byte `gorm:"column:icon"`
	Lbp1UsedSlots             int64  `gorm:"column:lbp1UsedSlots"`
	EntitledSlots             int64  `gorm:"column:entitledSlots"`
	FreeSlots                 int64  `gorm:"column:freeSlots"`
	CrossControlUsedSlots     int64  `gorm:"column:crossControlUsedSlots"`
	CrossControlEntitledSlots int64  `gorm:"column:crossControlEntitledSlots"`
	CrossControlFreeSlots     int64  `gorm:"column:crossControlFreeSlots"`
	LBP2UsedSlots             int64  `gorm:"column:lbp2UsedSlots"`
	LBP2EntitledSlots         int64  `gorm:"column:lbp2EntitledSlots"`
	LBP2FreeSlots             int64  `gorm:"column:lbp2FreeSlots"`
	LBP3UsedSlots             int64  `gorm:"column:lbp3UsedSlots"`
	LBP3EntitledSlots         int64  `gorm:"column:lbp3EntitledSlots"`
	LBP3FreeSlots             int64  `gorm:"column:lbp3FreeSlots"`
	HeartCount                int64  `gorm:"column:heartCount"`
	Planets                   []byte `gorm:"column:planets"`
	Yay2                      []byte `gorm:"column:yay2"`
	Boo2                      []byte `gorm:"column:boo2"`
	LocationX                 uint16 `gorm:"column:locationX"`
	LocationY                 uint16 `gorm:"column:locationY"`
	CommentsEnabled           bool   `gorm:"column:commentsEnabled"`
	Pins0                     uint32 `gorm:"column:pins0"`
	Pins1                     uint32 `gorm:"column:pins1"`
	Pins2                     uint32 `gorm:"column:pins2"`
}

func (*ImportUser) TableName() string {
	return "user"
}
