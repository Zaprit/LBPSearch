package model

import (
	"encoding/hex"
	"time"
)

type Slot struct {
	ID       uint32 `gorm:"column:id"`
	NpHandle string `gorm:"column:np_handle"`
	Author   *User  `gorm:"-"`

	LocationX           int64   `gorm:"column:location_x"`
	LocationY           int64   `gorm:"column:location_y"`
	Game                uint8   `gorm:"column:game"`
	Name                string  `gorm:"column:name"`
	Description         string  `gorm:"column:description"`
	RootLevel           []byte  `gorm:"column:root_level"`
	Icon                []byte  `gorm:"column:icon"`
	InitiallyLocked     bool    `gorm:"column:initially_locked"`
	IsSubLevel          bool    `gorm:"column:is_sub_level"`
	IsLBP1Only          bool    `gorm:"column:is_lbp1_only"`
	Background          int64   `gorm:"column:background"`
	Thumbnail           []byte  `gorm:"column:thumbnail"`
	Shareable           bool    `gorm:"column:shareable"`
	AuthorLabels        []byte  `gorm:"column:author_labels"`
	Leveltype           string  `gorm:"column:leveltype"`
	MinPlayers          uint8   `gorm:"column:min_players"`
	MaxPlayers          uint8   `gorm:"column:max_players"`
	IsAdventurePlanet   bool    `gorm:"column:is_adventure_planet"`
	PS4Only             bool    `gorm:"column:ps4_only"`
	HeartCount          int64   `gorm:"column:heart_count"`
	ThumbsUp            int64   `gorm:"column:thumbs_up"`
	ThumbsDown          int64   `gorm:"column:thumbs_down"`
	AverageRating       float64 `gorm:"column:average_rating"`
	MMPick              bool    `gorm:"column:mm_pick"`
	ReviewCount         int64   `gorm:"column:review_count"`
	CommentCount        int64   `gorm:"column:comment_count"`
	PhotoCount          int64   `gorm:"column:photo_count"`
	AuthorPhotoCount    int64   `gorm:"column:author_photo_count"`
	Tags                []byte  `gorm:"column:tags"`
	Labels              []byte  `gorm:"column:labels"`
	FirstPublished      uint64  `gorm:"column:first_published"`
	LastUpdated         uint64  `gorm:"column:last_updated"`
	Genre               string  `gorm:"column:genre"`
	CommentsEnabled     bool    `gorm:"column:comments_enabled"`
	ReviewsEnabled      bool    `gorm:"column:reviews_enabled"`
	PublishedIn         string  `gorm:"column:published_in"`
	PlayCount           int64   `gorm:"column:play_count"`
	CompletionCount     int64   `gorm:"column:completion_count"`
	LBP1PlayCount       int64   `gorm:"column:lbp1_play_count"`
	LBP1CompletionCount int64   `gorm:"column:lbp1_completion_count"`
	LBP1UniquePlayCount int64   `gorm:"column:lbp1_unique_play_count"`
	LBP2PlayCount       int64   `gorm:"column:lbp2_play_count"`
	LBP2CompletionCount int64   `gorm:"column:lbp2_completion_count"`
	UniquePlayCount     int64   `gorm:"column:unique_play_count"`
	LBP3PlayCount       int64   `gorm:"column:lbp3_play_count"`
	LBP3CompletionCount int64   `gorm:"column:lbp3_completion_count"`
	LBP3UniquePlayCount int64   `gorm:"column:lbp3_unique_play_count"`
	Got                 uint8   `gorm:"column:got"`

	Deleted      bool
	DeletedAt    *time.Time
	DeleteReason string

	DownloadCount int64
}

func (s *Slot) GetRootLevel() string {
	return hex.EncodeToString(s.RootLevel)
}

func (s *Slot) GetIcon() string {
	return hex.EncodeToString(s.Icon)
}

func (s *Slot) GetName() string {
	if s.Name == "" {
		return "Unnamed Level"
	}
	return s.Name
}

func (s *Slot) GetPublishedIn() string {
	if s.PublishedIn == "" {
		switch s.Game {
		case 0:
			return "LittleBigPlanet"
		case 1:
			return "LittleBigPlanet 2"
		case 2:
			if s.PS4Only {
				return "LittleBigPlanet 3 PS4/PS5"
			}
			return "LittleBigPlanet 3"
		}
	} else {
		switch s.PublishedIn {
		case "lbp2":
			return "LittleBigPlanet 2"
		case "lbp3ps3":
			return "LittleBigPlanet 3 PS3"
		case "lbp3ps4":
			return "LittleBigPlanet 3 PS4/PS5"
		}
	}
	return "Unknown"
}

func (s *Slot) GetPublished() string {
	return time.UnixMilli(int64(s.FirstPublished)).Format(time.DateTime)
}

func (s *Slot) GetUpdated() string {
	return time.UnixMilli(int64(s.LastUpdated)).Format(time.DateTime)
}
